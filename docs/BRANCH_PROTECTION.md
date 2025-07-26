# Branch Protection Rules

このドキュメントは、urlmapプロジェクトのブランチ保護ルールの推奨設定を説明します。

## mainブランチの保護設定

GitHub リポジトリの Settings > Branches から以下の設定を行ってください：

### 1. Branch protection rules を追加

- **Branch name pattern**: `main`

### 2. Protect matching branches の設定

#### ✅ Require a pull request before merging
- ✅ **Require approvals**: 1
- ✅ **Dismiss stale pull request approvals when new commits are pushed**
- ✅ **Require review from CODEOWNERS** (CODEOWNERSファイルがある場合)

#### ✅ Require status checks to pass before merging
- ✅ **Require branches to be up to date before merging**
- **Required status checks**:
  - `Test (1.23)`
  - `Test (1.24)`
  - `Compatibility Check`
  - `Lint`

#### ✅ Require conversation resolution before merging
すべてのPRコメントが解決されるまでマージを防ぎます。

#### ✅ Require linear history
マージコミットを禁止し、クリーンな履歴を保ちます。

#### ✅ Include administrators
管理者もこれらのルールに従う必要があります。

#### ✅ Restrict who can push to matching branches
必要に応じて、直接プッシュできるユーザーを制限します。

### 3. Rules applied to everyone including administrators

#### ✅ Do not allow bypassing the above settings
管理者を含むすべてのユーザーが上記の設定をバイパスできないようにします。

## 設定後の動作

1. **PRなしでの直接プッシュ不可**: mainブランチへの直接プッシュはブロックされます
2. **CIチェック必須**: すべてのCIチェックが通らないとマージできません
3. **レビュー必須**: 最低1人のレビュー承認が必要です
4. **最新状態の維持**: マージ前にmainブランチの最新変更を取り込む必要があります
5. **テスト失敗時のマージ防止**: テストが失敗している場合、マージボタンが無効になります

## トラブルシューティング

### 緊急時の対応
本当に緊急の修正が必要な場合：
1. hotfixブランチを作成
2. 修正を実装（可能な限りテストも含める）
3. 通常のPRプロセスを経由
4. レビュアーに緊急である旨を伝える

### CIが誤って失敗する場合
1. Re-run failed jobsボタンで再実行
2. 一時的な問題の場合は、管理者に相談
3. CI設定自体に問題がある場合は、CI設定を修正するPRを作成

## メンテナンス

- 定期的にRequired status checksを見直し、新しいCIジョブを追加
- チームの成長に応じてRequired approvalsの数を調整
- 必要に応じてCODEOWNERSファイルを更新