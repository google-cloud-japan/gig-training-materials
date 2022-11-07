### Hi, I am WIP.
[元ネタ](https://cloud.google.com/architecture/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke?hl=ja)


# [GIG ハンズオン] **Cloud Code、Cloud Build、Google Cloud Deploy、GKE を使用したアプリの開発と配信**

## 目次 - Table of Contents
- [Google Cloud プロジェクトの選択](##Google-Cloud-プロジェクトの選択)
- [[解説] ハンズオンの内容と目的](##[解説]-ハンズオンの内容と目的)
- [アーキテクチャの概要](##アーキテクチャの概要)
- [目標](##目標)
- [費用](##費用)
- [環境を準備する](###HI,-I-am-WIP)
  - [権限を設定する](###HI,-I-am-WIP)
  - [GKEクラスタを作成する](###HI,-I-am-WIP)
  - [IDEを開いてリポジトリのクローンを作成する](###HI,-I-am-WIP)
  - [ソースコード用のリポジトリとコンテナ用のリポジトリを作成する](###HI,-I-am-WIP)
- [CI / CD パイプラインを構成する](###HI,-I-am-WIP)
- [デベロッパー ワークスペース内でアプリケーションを変更する](###HI,-I-am-WIP)
  - [アプリケーションをビルドし、テストして、実行する](###HI,-I-am-WIP)
  - [変更する](###HI,-I-am-WIP)
  - [コードを commit する](###HI,-I-am-WIP)
- [本番環境に変更をデプロイする](###HI,-I-am-WIP)
  - [CI / CD パイプラインを開始してステージング環境にデプロイする](###HI,-I-am-WIP)
  - [リリースを本番環境に昇格させる](###HI,-I-am-WIP)
- [クリーンアップ](###HI,-I-am-WIP)
  - [オプション1: プロジェクトを削除する](###HI,-I-am-WIP)
- [オプション2: 個々のリソースを削除する](###HI,-I-am-WIP)
- [次のステップ](###HI,-I-am-WIP)

## Google Cloud プロジェクトの選択

1. ハンズオンを行う Google Cloud プロジェクトを作成し、 Google Cloud プロジェクトを選択して **Start/開始** をクリックしてください。

**なるべく新しいプロジェクトを作成してください。**

<walkthrough-project-setup>
</walkthrough-project-setup>

2. Cloud プロジェクトに対して課金が有効になっていることを確認します。詳しくは、[プロジェクトで課金が有効になっているかどうかを確認する方法](https://cloud.google.com/billing/docs/how-to/verify-billing-enabled?hl=ja)をご覧ください。

3. Artifact Registry, Cloud Build, Google Cloud Deploy, Cloud Source Repositories, Google Kubernetes Engine, Resource Manager, and Service Networking API を有効にします。

    [API を有効にするリンク](https://console.cloud.google.com/flows/enableapi?apiid=artifactregistry.googleapis.com%2Ccloudbuild.googleapis.com%2Cclouddeploy.googleapis.com%2Csourcerepo.googleapis.com%2Ccontainer.googleapis.com%2C+cloudresourcemanager.googleapis.com%2Cservicenetworking.googleapis.com&%3Bredirect=https%3A%2F%2Fconsole.cloud.google.com&hl=ja&_ga=2.152803194.2113702237.1667787342-853816604.1666918848)

4. Google Cloud コンソールで、「Cloud Shell をアクティブにする」をクリックします。

    [Cloud Shell をアクティブにする](https://console.cloud.google.com/?cloudshell=true&hl=ja&_ga=2.144924279.2113702237.1667787342-853816604.1666918848)

    Google Cloud コンソールの下部で [Cloud Shell](https://cloud.google.com/shell/docs/how-cloud-shell-works?hl=ja) セッションが開始し、コマンドライン プロンプトが表示されます。Cloud Shell はシェル環境です。Google Cloud CLI がすでにインストールされており、現在のプロジェクトの値もすでに設定されています。セッションが初期化されるまで数秒かかることがあります。

## [解説] **ハンズオンの内容と目的**

このチュートリアルでは、Google Cloud ツールの統合セットを使用して、開発、継続的インテグレーション(CI)、継続的デリバリー(CD)システムを設定し、使用する方法について説明します。このシステムを使用すると、アプリケーションを開発し、[Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine?hl=ja)にデプロイできます。

このチュートリアルは、ソフトウェア デベロッパーとオペレーターの両方を対象としており、完了すると両方の役割を果たせます。まず、CI / CD パイプラインを設定するオペレーターの役割を果たします。このパイプラインの主要なコンポーネントは、[Cloud Build](https://cloud.google.com/build?hl=ja)、[Artifact Registry](https://cloud.google.com/artifact-registry?hl=ja)、[Google Cloud Deploy](https://cloud.google.com/deploy?hl=ja) です。

次にデベロッパーとして機能し、[Cloud Code](https://cloud.google.com/code?hl=ja) を使用してアプリケーションを変更します。デベロッパーとして機能すると、このパイプラインが提供する統合されたエクスペリエンスが表示されます。

最後に、オペレータとして機能して、アプリケーションを本番環境にデプロイする手順を実行します。

このチュートリアルは、Google Cloud での gcloud コマンドの実行と GKE へのアプリケーション コンテナのデプロイに精通していることを前提としています。

この統合システムの主な特徴は次のとおりです。

- **開発とデプロイが迅速になります。**

  デベロッパー ワークスペースで変更を検証できるため、開発ループが効率的です。自動化された CI/CD システムと環境間のパリティの向上により、本番環境に変更を展開する際に、より多くの問題を検出できるため、デプロイが高速化されます。

- **開発、ステージング、本番環境全体で一貫性の向上によるメリットを享受できます。**

  このシステムのコンポーネントは、共通の Google Cloud ツールセットを使用します。

- **さまざまな環境で構成を再利用します。**

  こうした再利用は Skaffold を使用して行われますが、Skaffold は異なる環境に対して共通の構成形式を認めます。また、デベロッパーとオペレーターが同じ構成を更新して使用することもできます。

- **ワークフローの早い段階でガバナンスを適用します。**

  このシステムでは、本番環境、CI システム、開発環境でガバナンスに関する検証テストが適用されます。開発環境では、ガバナンスを適用することで、問題を検出し、早期に修正できます。

- **独自のツールでソフトウェア配信を管理します。**

  継続的デリバリーはフルマネージドであり、CD パイプラインのステージをレンダリングやデプロイの細部から分離します。

## **アーキテクチャの概要**

次の図に、このチュートリアルで使用するリソースを示します。

![](https://raw.githubusercontent.com/google-cloud-japan/gig-training-materials/16-gig-5-3-%E3%82%B3%E3%83%B3%E3%83%86%E3%83%B3%E3%83%84%E4%BD%9C%E6%88%90/gig05-03/img/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke.png)

このパイプラインを構成する 3 つの主要コンポーネントは次のとおりです。

1. 開発ワークスペースとしての **Cloud Code**

    このワークスペースの一部として、[minikube](https://minikube.sigs.k8s.io/docs/) で実行される開発クラスタで変更を確認できます。[Cloud Shell](https://cloud.google.com/shell?hl=ja) で Cloud Code と minikube クラスタを実行します。Cloud Shell は、ブラウザからアクセスできるオンライン開発環境です。コンピューティング リソース、メモリ、統合開発環境(IDE)を備え、Cloud Code もインストールされます。

2. アプリケーションのビルドとテストを行うための **Cloud Build**(パイプラインの「CI」部分)

    パイプラインのこの部分には、次のアクションが含まれます。

    - Cloud Build は、Cloud Build トリガーを使用して、ソース リポジトリに対する変更をモニタリングします。'
    - メインブランチに変更が commit されると、Cloud Build トリガーは次の処理を行います。
      - アプリケーション コンテナを再構築します。
      - ビルド アーティファクトを Cloud Storage バケットに配置します。
      - アプリケーション コンテナを Artifact Registry に配置します。
      - コンテナでテストを実行します。
      - Google Cloud Deploy を呼び出して、コンテナをステージング環境にデプロイします。このチュートリアルでは、ステージング環境は Google Kubernetes Engine クラスタです。
    - ビルドとテストが成功したら、Google Cloud Deploy を使用して、ステージング環境から本番環境にコンテナを昇格できます。

3. デプロイを管理するための **Google Cloud Deploy**(パイプラインの「CD」部分)

    パイプラインのこの部分では、Google Cloud Deploy が次の処理を行います。

    - [配信パイプライン](https://cloud.google.com/deploy/docs/terminology?hl=ja#delivery_pipeline)と[ターゲット](https://cloud.google.com/deploy/docs/terminology?hl=ja#target)を登録します。ターゲットはステージング クラスタと本番環境クラスタを表します。
    - Cloud Storage バケットを作成し、Skaffold レンダリング ソースとレンダリングされたマニフェストを作成したバケットに保存します。
    - ソースコードを変更するたびに(新しいリリース)(https://cloud.google.com/deploy/docs/terminology?hl=ja#release)を行います。このチュートリアルには 1 つの変更点があるため、新しいリリースを 1 つ行います。
    - アプリケーションを本番環境にデプロイします。この本番環境へのデプロイでは、運用担当者(または指定した人)がデプロイを手動で承認します。このチュートリアルでは、本番環境は Google Kubernetes Engine クラスタです。

Kubernetes ネイティブ アプリケーションの継続的な開発を容易にするコマンドライン ツールである [Skaffold](https://skaffold.dev/) は、これらのコンポーネントの基盤であり、開発、ステージング、本番環境の間で構成を共有できます。

Google Cloud はアプリケーションのソースコードを GitHub に保存します。このチュートリアルの一環として、このリポジトリのクローンを Cloud Source Repositories に作成して、CI / CD パイプラインに接続します。

このチュートリアルでは、システムのほとんどのコンポーネントで Google Cloud プロダクトを使用し、Skaffold でシステムを統合しています。Skaffold はオープンソースであるため、このような原則で Google Cloud、社内コンポーネント、サードパーティ コンポーネントを組み合わせて同様のシステムを作成できます。このソリューションはモジュール方式を採用しているため、開発パイプラインとデプロイ パイプラインの一部として段階的に導入できます。

## **目標**

オペレーターとして、次の操作を行います。

- CI パイプラインと CD パイプラインを設定します。この設定には次のものが含まれます。
  - 必要な権限を設定する。
  - ステージング環境と本番環境用の GKE クラスタを作成する。
  - ソースコード用のリポジトリを Cloud Source Repositories に作成する。
  - アプリケーション コンテナ用のリポジトリを Artifact Registry に作成する。
  - メインの GitHub リポジトリに Cloud Build トリガーを作成する。
  - Google Cloud Deploy デリバリー パイプラインとターゲットを作成する。ターゲットはステージング環境と本番環境です。
- ステージング環境にデプロイする CI / CD プロセスを開始し、本番環境に昇格させる。

開発者として、アプリケーションに変更を加えます。手順は次のとおりです。

- 事前構成された開発環境と連携するように、リポジトリのクローンを作成する。
- デベロッパー ワークスペース内でアプリケーションを変更する
- 変更をビルドおよびテストします。テストには、ガバナンスの検証テストが含まれます。
- 開発クラスタで変更を表示して検証します。このクラスタは minikube で実行されます。
- メイン リポジトリに変更を commit する。

## **費用**

このチュートリアルでは、課金対象である次の Google Cloud コンポーネントを使用します。

[Cloud Build](https://cloud.google.com/build/pricing?hl=ja)
[Google Cloud Deploy](https://cloud.google.com/deploy/pricing?hl=ja)
[Artifact Registry](https://cloud.google.com/artifact-registry/pricing?hl=ja)
[Google Kubernetes Engine](https://cloud.google.com/kubernetes-engine/pricing?hl=ja)
[Cloud Source Repositories](https://cloud.google.com/source-repositories/pricing?hl=ja)
[Cloud Storage](https://cloud.google.com/storage/pricing?hl=ja)
[料金計算ツール](https://cloud.google.com/products/calculator?hl=ja)を使うと、予想使用量に基づいて費用の見積もりを生成できます。

このチュートリアルを終了した後、作成したリソースを削除すると、それ以上の請求は発生しません。詳細については、[クリーンアップ](https://cloud.google.com/architecture/app-development-and-delivery-with-cloud-code-gcb-cd-and-gke?hl=ja#clean-up)をご覧ください。

