#!/bin/bash

# GitHub发布脚本 - 将多个commit压缩成单个版本commit推送到GitHub
# 用法: ./scripts/release-to-github.sh "v1.1.0" "Release description"

if [ $# -ne 2 ]; then
    echo "用法: $0 <版本号> <发布说明>"
    echo "例如: $0 'v1.1.0' 'Release v1.1.0: 新功能和bug修复'"
    exit 1
fi

VERSION=$1
MESSAGE=$2
BRANCH_NAME="github-release-temp"

echo "🚀 准备发布到GitHub: $VERSION"

# 检查是否有未提交的更改
if [[ -n $(git status --porcelain) ]]; then
    echo "❌ 错误: 存在未提交的更改，请先提交到本地仓库"
    exit 1
fi

# 创建临时分支
echo "📝 创建发布分支..."
git checkout -b $BRANCH_NAME

# 获取当前分支的第一个commit (GitHub上的最后一个commit)
LAST_GITHUB_COMMIT=$(git ls-remote github main | cut -f1)

if [ -n "$LAST_GITHUB_COMMIT" ]; then
    # 如果GitHub上有commit，则基于最后一个GitHub commit创建新的单commit
    echo "🔄 基于GitHub最后一个commit创建新版本..."
    git reset --soft $LAST_GITHUB_COMMIT
else
    # 如果是首次推送，压缩所有历史
    echo "🔄 压缩所有历史为单个commit..."
    FIRST_COMMIT=$(git rev-list --max-parents=0 HEAD)
    git reset --soft $FIRST_COMMIT
fi

# 创建新的发布commit
git add .
git commit -m "$MESSAGE"

# 推送到GitHub
echo "📤 推送到GitHub..."
git push github $BRANCH_NAME:main --force

# 清理临时分支
echo "🧹 清理临时分支..."
git checkout main
git branch -D $BRANCH_NAME

echo "✅ 成功发布 $VERSION 到GitHub!"
echo "🔗 查看: https://github.com/HoBeedzc/cc-switch"
