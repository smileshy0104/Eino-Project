@echo off
rem AI文档助手 Windows 安装脚本
rem 适用于 Windows 10/11 系统

echo ===============================
echo AI文档助手 Windows 安装脚本
echo ===============================

rem 检查管理员权限
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 请以管理员身份运行此脚本
    echo 右键点击cmd并选择"以管理员身份运行"
    pause
    exit /b 1
)

echo [1/8] 检查系统环境...

rem 检查Docker
docker --version >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 未检测到Docker
    echo 请先安装Docker Desktop: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

rem 检查Docker Compose
docker-compose --version >nul 2>&1
if %errorLevel% neq 0 (
    echo 错误: 未检测到Docker Compose
    echo 请确保Docker Desktop已正确安装
    pause
    exit /b 1
)

echo ✓ Docker 环境正常

echo [2/8] 创建必要目录...
if not exist logs mkdir logs
if not exist uploads mkdir uploads
if not exist backups mkdir backups
if not exist data mkdir data
if not exist data\mysql mkdir data\mysql
if not exist data\milvus mkdir data\milvus
if not exist data\redis mkdir data\redis

echo ✓ 目录创建完成

echo [3/8] 检查配置文件...
if not exist config\app.yaml (
    echo 警告: config\app.yaml 不存在
    echo 请确保配置文件存在并已正确配置
)

echo [4/8] 设置环境变量...
rem 设置默认环境变量（如果未设置）
if not defined AI_DOC_AI_API_KEY (
    echo 警告: AI_DOC_AI_API_KEY 环境变量未设置
    echo 请设置您的API密钥: set AI_DOC_AI_API_KEY=your-api-key
)

echo [5/8] 拉取Docker镜像...
docker-compose pull
if %errorLevel% neq 0 (
    echo 警告: 镜像拉取失败，将在启动时自动构建
)

echo [6/8] 启动服务...
docker-compose up -d
if %errorLevel% neq 0 (
    echo 错误: 服务启动失败
    pause
    exit /b 1
)

echo [7/8] 等待服务就绪...
timeout /t 30 /nobreak >nul

echo [8/8] 验证服务状态...
docker-compose ps

echo.
echo ===============================
echo ✅ 安装完成！
echo ===============================
echo.
echo 🌟 服务访问地址:
echo   • Web界面: http://localhost:8080
echo   • API文档: http://localhost:8080/swagger/index.html  
echo   • 健康检查: http://localhost:8080/health
echo.
echo 🔧 管理命令:
echo   • 查看状态: make status
echo   • 查看日志: make logs
echo   • 停止服务: make stop
echo   • 重启服务: make restart
echo.
echo 📝 日志文件位置: logs\app.log
echo.

rem 询问是否打开浏览器
set /p choice="是否现在打开Web界面? (Y/N): "
if /i "%choice%"=="Y" (
    start http://localhost:8080
)

pause