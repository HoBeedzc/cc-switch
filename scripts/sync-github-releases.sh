#!/bin/bash

# 双分支管理脚本：维护origin/main和github/main的同步
# origin/main: 完整开发历史
# github/main: 仅包含tag版本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否有未提交的更改
check_clean_working_tree() {
    if ! git diff-index --quiet HEAD --; then
        log_error "工作目录不干净，请先提交或暂存更改"
        exit 1
    fi
}

# 重建github/main分支，只包含tag版本
rebuild_github_main() {
    log_info "开始重建github/main分支..."
    
    # 1. 检查工作目录
    check_clean_working_tree
    
    # 2. 获取所有tag并按版本排序
    log_info "获取所有tag版本..."
    tags=$(git tag -l --sort=version:refname)
    
    if [ -z "$tags" ]; then
        log_error "未找到任何tag"
        exit 1
    fi
    
    # 3. 创建新的orphan分支
    log_info "创建临时分支..."
    git checkout --orphan temp-github-main
    git rm -rf . > /dev/null 2>&1 || true
    
    # 4. 按顺序应用每个tag的内容
    for tag in $tags; do
        log_info "处理tag: $tag"
        
        # 获取tag对应的commit
        tag_commit=$(git rev-list -n 1 "$tag")
        
        # 检出该commit的内容
        git checkout "$tag_commit" -- . > /dev/null 2>&1 || true
        
        # 创建干净的提交
        git add .
        commit_msg="Release $tag"
        
        # 如果有现有的tag消息，使用它
        if git tag -l -n1 "$tag" | grep -q " "; then
            tag_msg=$(git tag -l -n1 "$tag" | cut -d' ' -f2-)
            commit_msg="$commit_msg: $tag_msg"
        fi
        
        git commit -m "$commit_msg" > /dev/null 2>&1 || true
        log_success "已添加 $tag"
    done
    
    # 5. 替换原github/main分支
    log_info "更新github/main分支..."
    git branch -D github-main-backup > /dev/null 2>&1 || true
    git branch -m main github-main-backup > /dev/null 2>&1 || true
    git branch -m temp-github-main main
    
    log_success "github/main分支重建完成"
}

# 同步到github仓库
sync_to_github() {
    log_info "同步到github仓库..."
    
    # 强制推送新的main分支到github
    git push github main --force
    
    # 推送所有tags
    git push github --tags --force
    
    log_success "已同步到github仓库"
}

# 恢复origin/main分支
restore_origin_main() {
    log_info "恢复origin/main工作分支..."
    
    # 切换回origin/main的内容
    git checkout github-main-backup
    git branch -D main > /dev/null 2>&1 || true
    git branch -m github-main-backup main
    
    # 同步origin/main到最新
    git fetch origin
    git reset --hard origin/main
    
    log_success "已恢复origin/main分支"
}

# 主函数
main() {
    echo "=================================="
    echo "    双分支同步管理工具"
    echo "=================================="
    echo
    
    case "${1:-}" in
        "rebuild")
            rebuild_github_main
            sync_to_github
            restore_origin_main
            ;;
        "sync")
            sync_to_github
            ;;
        "status")
            log_info "检查分支状态..."
            echo
            echo "当前分支: $(git branch --show-current)"
            echo "Origin状态:"
            git log origin/main --oneline -5
            echo
            echo "GitHub状态:"
            git log github/main --oneline -5 2>/dev/null || echo "github/main分支不存在"
            echo
            echo "Tags:"
            git tag -l --sort=-version:refname
            ;;
        *)
            echo "用法: $0 {rebuild|sync|status}"
            echo
            echo "命令说明:"
            echo "  rebuild  - 重建github/main分支，只包含tag版本"
            echo "  sync     - 同步当前分支到github仓库"
            echo "  status   - 查看分支状态"
            echo
            exit 1
            ;;
    esac
}

main "$@"
