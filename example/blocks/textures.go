package blocks

import (
	"github.com/hajimehoshi/go-ebiten/graphics"
	"image"
	_ "image/png"
	"os"
	"sync"
)

type namePath struct {
	name string
	path string
}

type nameSize struct {
	name string
	size Size
}

type Textures struct {
	textureFactory    graphics.TextureFactory
	texturePaths      chan namePath
	renderTargetSizes chan nameSize
	textures          map[string]graphics.TextureId
	renderTargets     map[string]graphics.RenderTargetId
	sync.RWMutex
}

func NewTextures(textureFactory graphics.TextureFactory) *Textures {
	textures := &Textures{
		textureFactory:    textureFactory,
		texturePaths:      make(chan namePath),
		renderTargetSizes: make(chan nameSize),
		textures:          map[string]graphics.TextureId{},
		renderTargets:     map[string]graphics.RenderTargetId{},
	}
	go textures.loop()
	return textures
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (t *Textures) loop() {
	for {
		select {
		case p := <-t.texturePaths:
			name := p.name
			path := p.path
			go func() {
				img, err := loadImage(path)
				if err != nil {
					panic(err)
				}
				ch := t.textureFactory.CreateTexture(
					img,
					graphics.FilterNearest)
				go func() {
					e := <-ch
					if e.Error != nil {
						panic(e.Error)
					}
					t.Lock()
					defer t.Unlock()
					t.textures[name] = e.Id
				}()
			}()
		case s := <-t.renderTargetSizes:
			name := s.name
			size := s.size
			go func() {
				ch := t.textureFactory.CreateRenderTarget(
					size.Width,
					size.Height,
					graphics.FilterNearest)
				go func() {
					e := <-ch
					if e.Error != nil {
						panic(e.Error)
					}
					t.Lock()
					defer t.Unlock()
					t.renderTargets[name] = e.Id
				}()
			}()
		}
	}
}

func (t *Textures) RequestTexture(name string, path string) {
	t.texturePaths <- namePath{name, path}
}

func (t *Textures) RequestRenderTarget(name string, size Size) {
	t.renderTargetSizes <- nameSize{name, size}
}

func (t *Textures) Has(name string) bool {
	t.RLock()
	defer t.RUnlock()
	_, ok := t.textures[name]
	if ok {
		return true
	}
	_, ok = t.renderTargets[name]
	return ok
}

func (t *Textures) GetTexture(name string) graphics.TextureId {
	t.RLock()
	defer t.RUnlock()
	return t.textures[name]
}

func (t *Textures) GetRenderTarget(name string) graphics.RenderTargetId {
	t.RLock()
	defer t.RUnlock()
	return t.renderTargets[name]
}