# 部署说明

当前生产部署采用 `nginx + systemd + 静态文件发布` 的方式。

## 结构

- 后端服务：`deploy/systemd/wuwa-echo-backend.service`
- 前端构建服务：`deploy/systemd/wuwa-echo-frontend.service`
- 前端发布脚本：`deploy/scripts/publish-frontend.sh`
- `nginx` 配置模板：`deploy/nginx/wuwa-echo.conf`

## 前端部署方式

前端不再常驻运行 `vite preview`。

现在的生产流程是：

1. 在 `frontend/` 目录执行 `npm run build-only`
2. 将 `frontend/dist/` 同步到 `/var/www/wuwa-echo`
3. 由 `nginx` 直接托管 `/var/www/wuwa-echo`

对应脚本就是：

```bash
/root/wuwa/echo/deploy/scripts/publish-frontend.sh
```

`systemd` 前端服务是 `oneshot` 类型，所以它的职责不是常驻提供 HTTP 服务，而是“构建并发布一次前端”：

```bash
systemctl start wuwa-echo-frontend.service
```

执行完成后，正常状态应当是：

- `active (exited)`

这表示构建发布成功并退出，不是故障。

## nginx 职责

`nginx` 负责两件事：

- 托管前端静态文件目录 `/var/www/wuwa-echo`
- 将 `/api/` 反代到后端 `http://127.0.0.1:8888/`

同时配置了 SPA fallback：

- `try_files $uri $uri/ /index.html;`

这样 Vue Router 的前端路由可以直接刷新访问。

## 后端部署方式

后端由 `systemd` 常驻运行：

- 工作目录：`/root/wuwa/echo/backend`
- 环境变量：读取 `backend/.env`
- 可执行文件：`/root/wuwa/echo/backend/bin/wuwa-echo-backend`

常用命令：

```bash
systemctl status wuwa-echo-backend.service
systemctl restart wuwa-echo-backend.service
```

## 典型发布步骤

更新后端：

```bash
cd /root/wuwa/echo/backend
./scripts/build-go.sh
systemctl restart wuwa-echo-backend.service
```

更新前端：

```bash
cd /root/wuwa/echo
systemctl start wuwa-echo-frontend.service
```

若 `nginx` 配置有改动，再执行：

```bash
nginx -t
systemctl reload nginx
```
