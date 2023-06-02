# GCP Handson materials for GIG 5-3 (CI/CD on Google Cloud)

[![Open in Cloud Shell](https://gstatic.com/cloudssh/images/open-btn.png)](https://ssh.cloud.google.com/cloudshell/open?cloudshell_git_repo=https://github.com/google-cloud-japan/gig-training-materials&cloudshell_working_dir=gig05-03&cloudshell_tutorial=tutorial.md&shellonly=true)

**This is not an officially supported Google product**.

see [tutorial.md](tutorial.md) for more details

---
### **Cloud Code、Cloud Build、Google Cloud Deploy、GKE を使用したアプリの開発と配信**
[オリジナル公式ドキュメント](https://cloud.google.com/architecture/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke?hl=ja)

<!-- TODO: Update link URL after the PR merged -->
--- **目次 - Table of Contents** ---
- [解説] ハンズオンの内容と目的
- アーキテクチャの概要
- 目標
- 費用
- Google Cloud プロジェクトの選択
- 環境を準備する
  - 権限を設定する
  - GKEクラスタを作成する
  - IDEを開いてリポジトリのクローンを作成する
  - ソースコード用のリポジトリとコンテナ用のリポジトリを作成する
- CI / CD パイプラインを構成する
- デベロッパー ワークスペース内でアプリケーションを変更する
  - アプリケーションをビルドし、テストして、実行する
  - 変更する
  - コードを commit する
- 本番環境に変更をデプロイする
  - CI / CD パイプラインを開始してステージング環境にデプロイする
  - リリースを本番環境に昇格させる
- クリーンアップ
  - オプション1: プロジェクトを削除する
  - オプション2: 個々のリソースを削除する
- 次のステップ
