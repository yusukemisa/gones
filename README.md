# gones
Go言語によるファミコンエミュレーター実装

# Test
Run gones
```
 make run
```

Run test
```
 make test
```

# Road map

## 1. Run Hello World Rom
Hello World Rom is a sample rom distributed [here](http://hp.vector.co.jp/authors/VA042397/nes/sample.html).

- [x] Romファイルからプログラム部とキャラクター部に分けて読み込む
- [x] Romをメモリ上にロードして命令セット部分をとりだすfetchの実装
- [x] fetchしたcodeをdecodeしInstructionに変換する仕組み
- [x] Hello World Romを10cycle読んで命令を実行
- [x] Hello World Romを20cycle読んで命令を実行
- [x] Hello World Romの動作に必要な命令セットの実装する
- [x] 画面描画ライブラリ(go-sdl2)導入
- [x] PPU実装
- [x] Hello world ROM実行
  - <img width="291" alt="image" src="https://user-images.githubusercontent.com/25860926/212910341-49448c42-e7f2-41c8-9378-31dc0842403c.png">


- [x] 背景色表示
- [ ] Bus実装
- [ ] ジョイパッド実装

参考
- [ファミコンエミュレータの創り方　- Hello, World!編 -](https://qiita.com/bokuweb/items/1575337bef44ae82f4d3)
- [NES研究室](http://hp.vector.co.jp/authors/VA042397/nes/6502.html)
- [6502 Instruction Reference](https://www.nesdev.org/obelisk-6502-guide/reference.html)
- [Writing NES Emulator in Rust](https://bugzmanov.github.io/nes_ebook/)
- [Writing NES Emulator in Rustをやった](https://zenn.dev/razokulover/articles/1191ca55f9f22e)


## 1.1 sdl2で画面描画する
https://github.com/veandco/go-sdl2

ebitengineはいつ画面の更新を行うかがよくわからなかったので使用を断念
https://ebitengine.org/ja/
