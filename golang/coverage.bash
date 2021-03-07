#!/bin/bash

# 実行時ディレクトリ（htmlが保存されるパス）
SCRIPT_DIR=$(cd $(dirname $0); pwd)

# 実行コマンド
# bash bash coverage.bash [パッケージ] [テストファイル名]
work="cover_work"
output="coverage"

# エラーハンドリング
if [ -z $1 ]; then
    echo "パッケージ名を指定してください"
    exit 1
elif [ -z $2 ]; then
    echo "テストファイル名を指定してください"
    exit 1
fi

# テストしやすい様に移動
cd $1

go test -coverprofile=${work}.out
echo mode: set > ${output}.out
cat ${work}.out | grep ${2%_test.go} >> ${output}.out

go tool cover -html=${output}.out -o $SCRIPT_DIR/${output}.html

# 不要なファイルの削除
rm ${work}.out
rm ${output}.out


