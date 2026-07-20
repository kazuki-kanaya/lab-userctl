# lab-userctl

![Platform: Ubuntu](https://img.shields.io/badge/platform-Ubuntu-E95420?logo=ubuntu&logoColor=white)

[English README](README.md)

Ubuntuサーバー上のローカルLinuxユーザーを、対話形式で登録するシンプルなCLIツールです。

## 背景

研究室のGPUサーバーへ新規ユーザーを登録するとき、ユーザー作成、sudo権限の付与、SSH公開鍵の登録、所有者やパーミッションの設定を毎回手作業で行うのが面倒でした。

このツールは、そのような定型作業を手軽に行うために作っています。

## できること

1回のコマンドで、以下のアカウント設定作業をまとめて行えます。

- 必要に応じたローカルユーザー作成
- 新規ユーザーへのパスワード設定
- 任意のsudo権限付与
- 任意のSSH公開鍵登録
- `.ssh` と `authorized_keys` の安全な権限設定
- SSH公開鍵の重複登録防止

## 使い方

```bash
sudo lab-userctl register
```

ユーザー名、新規作成時のパスワード、sudo権限の付与、SSH公開鍵の登録を対話形式で確認します。

受け付けるのはSSH公開鍵だけです。秘密鍵は絶対に入力しないでください。秘密鍵を入力するとエラーになります。

## インストール

最新のLinux向けリリースをインストールします。

```bash
curl -fsSL https://raw.githubusercontent.com/kazuki-kanaya/lab-userctl/main/scripts/install.sh | sh
```

インストーラーは、公開済みのSHA-256チェックサムでダウンロードしたアーカイブを検証してから、`lab-userctl`を`/usr/local/bin`へインストールします。

## ビルド

```bash
goenv install 1.26.5
goenv use 1.26.5
go build -o lab-userctl .
```

システムユーザーを変更するツールなので、本番サーバーで使う前に破棄可能なUbuntu VMまたはテスト用アカウントで確認してください。
