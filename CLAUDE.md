# CLAUDE.md

## API Key 架构

后端**不配置全局 API Key**。不存在 `OPENAI_API_KEY` 环境变量，也没有 `openAIConfigured` 的概念。

每个用户通过自己的 apikey 登录后端，该 apikey 同时作为该用户的 OpenAI API Key 使用（后端存储为加密的 `ApikeyCipher`，调用 OpenAI 时解密使用）。

因此：
- 不要在前端添加"后端未配置 API Key"之类的提示
- 不要在后端添加 `OPENAI_API_KEY` 环境变量读取或 `OpenAIConfigured` 字段
- 前端的 `apiKey` 字段是用户登录凭证，不是后端配置
