# 自定义域名配置

OpenTMD 文档站点部署在 **GitHub Pages**，自定义域名：

- **主站**: [www.opentmd.com](https://www.opentmd.com)
- **根域**: [opentmd.com](https://opentmd.com)（需 DNS 配置）

构建产物由 `.github/workflows/pages.yml` 自动发布；`docs/site/CNAME` 指定 `www.opentmd.com`。

## 1. 启用 GitHub Pages

在仓库 **Settings → Pages**：

| 项 | 值 |
|----|-----|
| Source | **GitHub Actions** |
| Custom domain | `www.opentmd.com` |
| Enforce HTTPS | ✅ 启用 |

推送至 `master`/`main` 且 `docs/` 有变更时，workflow 自动构建并部署。

## 2. DNS 记录（域名注册商）

在 `opentmd.com` 的 DNS 控制台添加：

### www 子域（CNAME）

| 类型 | 主机 | 值 |
|------|------|-----|
| CNAME | `www` | `<org-or-user>.github.io` |

示例：若仓库为 `opentmd/opentmd-cli`，且 Pages 用户域为组织站，则 CNAME 指向组织 GitHub Pages 域名（以 GitHub Pages 设置页显示为准）。

### 根域 apex（A 记录）

| 类型 | 主机 | 值 |
|------|------|-----|
| A | `@` | `185.199.108.153` |
| A | `@` | `185.199.109.153` |
| A | `@` | `185.199.110.153` |
| A | `@` | `185.199.111.153` |

> GitHub Pages 官方 IP，用于 `opentmd.com` 裸域访问。部分注册商支持 ALIAS/ANAME 指向 `www.opentmd.com` 作为替代方案。

### 可选：根域跳转到 www

若希望 `opentmd.com` 统一跳转到 `www.opentmd.com`，可在 DNS 提供商配置 URL 重定向（非 GitHub 侧），或仅使用 apex A 记录由 GitHub 同时服务两域（在 Pages 自定义域中可同时验证）。

## 3. 验证

DNS 传播后（通常 5–30 分钟，最长 48 小时）：

```bash
dig www.opentmd.com CNAME +short
dig opentmd.com A +short
curl -I https://www.opentmd.com/
```

浏览器访问：

- https://www.opentmd.com/zh/index.html
- https://www.opentmd.com/ （自动跳转文档首页）

## 4. 本地预览

```bash
cd docs/site
npm install
npm run build
npm run serve
# 打开 http://localhost:4173
```

## 5. 文件说明

| 文件 | 说明 |
|------|------|
| `docs/site/nav.json` | 侧边栏导航与 markdown 源映射 |
| `docs/site/build.mjs` | Markdown → HTML 构建脚本 |
| `docs/site/CNAME` | GitHub Pages 自定义域名 |
| `docs/site/dist/` | 构建输出（gitignore，由 CI 生成） |

修改 `docs/**/*.md` 后推送即可更新线上文档；无需手动提交 `dist/`。
