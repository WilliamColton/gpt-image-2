---
phase: 05-cost-revenue-analytics
plan: 03
subsystem: backend-generation-billing
tags:
  - billing
  - failover
  - endpoint-attribution
  - generation
  - tdd
requires:
  - 05-01 (BillingRecord model, RecordBillingForSuccessfulImages)
  - 05-02 (endpoint cost pricing, GetSalePriceX10000)
provides:
  - Endpoint attribution on every GeneratedImage
  - Billing row creation after successful image save
  - Saved-success slice for correct metadata pairing
affects:
  - backend-go/service/openai.go
  - backend-go/handler/generate.go
tech-stack:
  added:
    - Go test isolation patterns (sqlite in-memory with AutoMigrate)
  patterns:
    - TDD (RED/GREEN/REFACTOR)
    - saved-success slice pattern replacing index-based pairing
    - endpoint attribution stamping in withFailover
key-files:
  created:
    - backend-go/service/openai_failover_test.go
    - backend-go/handler/generate_billing_test.go
  modified:
    - backend-go/service/openai.go
    - backend-go/handler/generate.go
decisions:
  - "Endpoint attribution stamped in withFailover only after successful endpoint call; failed attempts leave no trace"
  - "Billing uses saved-success slice, never requested n or API result count"
  - "User label passed to executeImageGeneration for billing snapshot"
  - "config.GetSalePriceX10000() called at billing time for immutable price snapshot"
  - "buildPerImageMetadata replaced with buildPerImageMetadataFromSaved to eliminate index-shift on partial save failure"
duration: "~12 minutes"
completed_date: "2026-05-23"
---

# Phase 05 Plan 03: Billing Integration in Generation Success Path

Plumb endpoint attribution through the OpenAI/failover result so every generated image carries the URL and unit cost of the endpoint that produced it. Replace the index-based image-save logic with a saved-success slice that correctly pairs output image IDs with their source generated images, then write immutable billing rows only for successfully saved images.

## Task Summary

| # | Task | Type | Commit | Tests |
|---|------|------|--------|-------|
| 1 | Carry successful endpoint attribution through OpenAI results | auto (TDD) | `02cc1f1` | 3 pass |
| 2 | Write billing after successful image save | auto (TDD) | `4b1ccea` | 5 pass |

## Task 1: Endpoint Attribution

**Implemented:** Two new fields on `GeneratedImage` (`EndpointBaseURL string`, `UnitCostX10000 int64`). `withFailover` stamps these fields from the successful endpoint's configuration only after a successful function call, and only when the fields are empty. Failed endpoints never stamp images. `mergeConcurrentResults` already preserved per-image fields by appending slices, so concurrent generation mixed-endpoint attribution is preserved.

**Tests created (3):**
- `TestWithFailoverStampsEndpointAttribution` — single successful endpoint stamps both fields
- `TestWithFailoverFailedFirstEndpointDoesNotStampAttribution` — failover from failed to success; attribution comes from the success endpoint, not the failed one
- `TestMergeConcurrentResultsPreservesEndpointAttribution` — concurrent merge preserves different endpoints' attribution per-image

## Task 2: Billing Integration

**Implemented:** New `savedGeneratedImage` struct pairs each saved output image ID with the generated image that produced it. `saveGeneratedImagesWithAttribution` replaces `saveGeneratedImages`, returning a slice where each entry is a (outputImageID, GeneratedImage) pair — partial save failures simply skip entries without shifting later ones.

`buildBillingInput` constructs a `BillingBatchInput` from the saved-success slice, using `config.GetSalePriceX10000()` for the immutable sale price snapshot. `RecordBillingForSuccessfulImages` is called only after all output image IDs are known and before the task is upserted as "done."

`buildPerImageMetadataFromSaved` replaces `buildPerImageMetadata`, building metadata maps from the saved-success slice instead of by index, eliminating the metadata-shift bug when earlier images fail to save.

`executeImageGeneration` signature changed to include `userLabel string`; `GenerateImage` passes `user.Label`.

**Ordering guarantee:** image save -> billing -> used_count increment -> task upsert. Billing and used_count both use the same saved-success count.

**Tests created (5):**
- `TestSaveGeneratedImagesWithAttribution_ReturnsCorrectPairings` — all three images save; each outputs correct attribution
- `TestSaveGeneratedImagesWithAttribution_PartialFailureDoesNotShiftPairings` — middle image save fails; first and third entries are correctly paired without shift
- `TestRecordBillingForSavedImages_RowCountEqualsSaveCount` — two saved images produce two billing rows; snapshots match
- `TestBuildPerImageMetadata_NoShiftOnPartialSave` — metadata maps built from saved-success slice; keys match output IDs, not original indices
- `TestBuildBillingInput_UsesConfigSalePriceSnapshot` — `config.GetSalePriceX10000()` value captured at billing time

## Verification

All targeted verification patterns pass:
- `cd backend-go && go test ./service/... -run 'Test.*Failover.*Attribution|Test.*Endpoint.*Attribution|Test.*Failover'` — 3/3 pass
- `cd backend-go && go test ./handler/... -run 'Test.*Billing|Test.*SavedGeneratedImages|Test.*PerImageMetadata'` — 3/3 pass (full suite: 5/5 pass including save tests)
- Full service + handler suites pass with no regressions

## Deviations from Plan

None — plan executed exactly as written.

## Threat Model Coverage

| Threat | Mitigation | Status |
|--------|-----------|--------|
| T-05-03-01 (Tampering: endpoint attribution) | Stamp only after success; tests for failover | Implemented |
| T-05-03-02 (Tampering: billing counts) | Saved-success slice after SaveDataURLImage; tests assert row count equals saved count | Implemented |
| T-05-03-03 (Repudiation: snapshot sources) | TaskID, UserID, UserLabel, EndpointBaseURL, UnitCost, UnitSale, CreatedAt all stored per-row | Implemented |

## Known Stubs

None. All billing data flows through the full path: endpoint attribution -> saved-success slice -> billing row creation.

## Self-Check: PASSED

- [x] `backend-go/service/openai.go` — modified, committed in `02cc1f1`
- [x] `backend-go/service/openai_failover_test.go` — created, committed in `02cc1f1`
- [x] `backend-go/handler/generate.go` — modified, committed in `4b1ccea`
- [x] `backend-go/handler/generate_billing_test.go` — created, committed in `4b1ccea`
- [x] All tests pass: `go test ./service/... ./handler/...`
