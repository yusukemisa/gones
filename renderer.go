package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
)

type Renderer struct {
	texture  *sdl.Texture
	renderer *sdl.Renderer
}

func (r *Renderer) renderScreen(screen *Screen) {
	// byte列から描画内容となるtextureを作成
	// rect   : 更新する領域. nilの場合テクスチャ全体が対象.
	// pixels : 生データ
	// pitch  : ピクセルデータの水平方向バイト数. ライン間のパディング含む
	if err := r.texture.UpdateRGBA(nil, screen.pixels, 800*4); err != nil {
		log.Fatal(err)
	}

	// rendererを一度クリア
	if err := r.renderer.Clear(); err != nil {
		log.Fatal(err)
	}
	// 描画内容であるtextureを描画対象であるrendererにコピー
	if err := r.renderer.Copy(r.texture, nil, nil); err != nil {
		log.Fatal(err)
	}
	// 画面更新
	r.renderer.Present()

	sdl.Delay(20)
}
