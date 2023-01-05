# gones
Go言語によるファミコンエミュレーター実装

# Road map

## 1. Run Hello World Rom
Hello World Rom is a sample rom distributed [here](http://hp.vector.co.jp/authors/VA042397/nes/sample.html).

- [x] Romファイルからプログラム部とキャラクター部に分けて読み込む
- [x] Romをメモリ上にロードして命令セット部分をとりだすfetchの実装
- [x] fetchしたcodeをdecodeしInstructionに変換する仕組み
- [x] Hello World Romを10cycle読んで命令を実行
- [ ] Hello World Romを20cycle読んで命令を実行
- [ ] Hello World Romの動作に必要な命令セットの実装する
- [ ] PPU実装
- [ ] PPUの処理結果をjpgで書き出す

参考
- [ファミコンエミュレータの創り方　- Hello, World!編 -](https://qiita.com/bokuweb/items/1575337bef44ae82f4d3)
- [NES研究室](http://hp.vector.co.jp/authors/VA042397/nes/6502.html)
- [6502 Instruction Set](https://www.masswerk.at/6502/6502_instruction_set.html#BNE)


## 1.1 EbitengineでHello World Romを描画する

https://ebitengine.org/ja/
