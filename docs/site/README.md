# OpenTMD 文档站点

GitHub Pages 静态文档，源文件为 `docs/**/*.md`，构建输出到 `dist/`。

## 本地开发

```bash
cd docs/site
npm install
npm run build
npm run serve   # http://localhost:4173
```

## 结构

```
docs/site/
├── nav.json          # 导航与 markdown 映射
├── build.mjs         # 构建脚本
├── assets/           # CSS / JS
├── CNAME             # www.opentmd.com
├── index.html        # 根路径跳转
└── dist/             # 构建产物（gitignore）
```

## 部署

推送到 `master`/`main` 后，`.github/workflows/pages.yml` 自动构建并发布。

自定义域名配置见 [DNS.md](DNS.md)。

## 添加页面

1. 在 `docs/` 下编写 markdown
2. 在 `nav.json` 中注册 `slug`、`title`、`src`
3. 推送后 CI 自动更新线上站点
