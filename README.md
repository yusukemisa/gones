# gones
Go言語によるファミコンエミュレーター実装

# Road map

## 1. Run Hello World Rom
Hello World Rom is a sample rom distributed [here](http://hp.vector.co.jp/authors/VA042397/nes/sample.html).

- [x] Romファイルからプログラム部とキャラクター部に分けて読み込む
- [ ] CPUのメモリマップにプログラムをロードする
- [ ] Romをメモリ上にロードして命令セット部分をとりだす
- [ ] Hello World Romの動作に必要な命令セットの実装する
- [ ] PPU実装
- [ ] PPUの処理結果をjpgで書き出す

参考
https://qiita.com/bokuweb/items/1575337bef44ae82f4d3

## 1.1 EbitengineでHello World Romを描画する

https://ebitengine.org/ja/
