/*
Connor Henderson Project 3 Game with web assembly support
*/
package main

import (
	"embed"
	"fmt"
	"github.com/blizzy78/ebitenui"
	"github.com/blizzy78/ebitenui/image"
	"github.com/blizzy78/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
	"time"
)

//go:embed assets/*
var EmbeddedAssets embed.FS

const (
	GameWidth   = 1400
	GameHeight  = 700
	PlayerSpeed = 5
)

var (
	mplusNormalFont font.Face
	mplusBigFont    font.Face
	textWidget      *widget.Text
	showUI          bool
	g               Game
)

type Sprite struct {
	pict *ebiten.Image
	xloc int
	yloc int
	dX   int
	dY   int
}

type Game struct {
	AppUI   *ebitenui.UI
	showUI  bool
	player  Sprite
	score   int
	enemy   []Sprite
	drawOps ebiten.DrawImageOptions
}

func main() {
	rand.Seed(int64(time.Now().Second()))
	ebiten.SetWindowSize(GameWidth, GameHeight)
	ebiten.SetWindowTitle("Golem Knight The IV")
	simpleGame := Game{score: 0}
	simpleGame.player = Sprite{
		pict: loadPNGImageFromEmbedded("golem-preview.png"),
		xloc: 200,
		yloc: 300,
		dX:   0,
		dY:   0,
	}
	simpleGame.enemy = simpleGame.fillEnemySlice(simpleGame.enemy)
	simpleGame.AppUI = MakeUIWindow()
	if err := ebiten.RunGame(&simpleGame); err != nil {
		log.Fatal("Oh no! something terrible happened and the game crashed", err)
	}
}

func (g *Game) Update() error {
	processPlayerInput(g)
	if g.score >= 10 && g.score%10 == 0 {
		g.AppUI.Update()
	}
	if g.showUI == false {
		g.AppUI.Update()
	}
	return nil
}

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	mplusBigFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	const x = 20
	screen.Fill(colornames.Crimson)
	msg := fmt.Sprintf("%d", g.score)
	g.drawOps.GeoM.Reset()
	g.drawOps.GeoM.Translate(float64(g.player.xloc), float64(g.player.yloc))
	screen.DrawImage(g.player.pict, &g.drawOps)
	for i := 0; i < len(g.enemy); i++ {
		g.drawOps.GeoM.Reset()
		g.drawOps.GeoM.Translate(float64(g.enemy[i].xloc), float64(g.enemy[i].yloc))
		screen.DrawImage(g.enemy[i].pict, &g.drawOps)
		if collision(g.player, g.enemy[i]) {
			g.score = g.score + 1
			g.enemy = remove(g.enemy, i)
		}
	}
	text.Draw(screen, msg, mplusNormalFont, x, 40, color.White)
	if g.score >= 10 && g.score%10 == 0 {
		g.showUI = true
	}
	if g.showUI == true {
		g.AppUI.Draw(screen)
	}
	if g.showUI == false {
		screen.Fill(colornames.Crimson)
		msg := fmt.Sprintf("%d", g.score)
		g.drawOps.GeoM.Reset()
		g.drawOps.GeoM.Translate(float64(g.player.xloc), float64(g.player.yloc))
		screen.DrawImage(g.player.pict, &g.drawOps)
		for i := 0; i < len(g.enemy); i++ {
			g.drawOps.GeoM.Reset()
			g.drawOps.GeoM.Translate(float64(g.enemy[i].xloc), float64(g.enemy[i].yloc))
			screen.DrawImage(g.enemy[i].pict, &g.drawOps)
			if collision(g.player, g.enemy[i]) {
				g.score = g.score + 1
				g.enemy = remove(g.enemy, i)
			}
		}
		text.Draw(screen, msg, mplusNormalFont, x, 40, color.White)
	}
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int) (*image.NineSlice, error) {
	i := loadPNGImageFromEmbedded(path)

	w, h := i.Size()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func MakeUIWindow() (GUIhandler *ebitenui.UI) {
	background := image.NewNineSliceColor(color.Gray16{})
	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{false, true, false}),
			widget.GridLayoutOpts.Padding(widget.Insets{
				Top:    20,
				Bottom: 20,
			}),
			widget.GridLayoutOpts.Spacing(0, 20))),
		widget.ContainerOpts.BackgroundImage(background))
	textInfo := widget.TextOptions{}.Text("Pheeeeew!! They're all dead. I really need to retire....", basicfont.Face7x13, color.White)

	idle, err := loadImageNineSlice("button-idle.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	hover, err := loadImageNineSlice("button-hover.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	pressed, err := loadImageNineSlice("button-pressed.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	disabled, err := loadImageNineSlice("button-disabled.png", 20, 0)
	if err != nil {
		log.Fatalln(err)
	}
	buttonImage := &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}
	button := widget.NewButton(
		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),
		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Exit the game", basicfont.Face7x13, &widget.ButtonTextColor{
			Idle: color.RGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:  30,
			Right: 30,
		}),
		widget.ButtonOpts.ClickedHandler(clickToQuit),
	)
	button2 := widget.NewButton(
		// specify the images to use
		widget.ButtonOpts.Image(buttonImage),
		// specify the button's text, the font face, and the color
		widget.ButtonOpts.Text("Exit the game", basicfont.Face7x13, &widget.ButtonTextColor{
			Idle: color.RGBA{0xdf, 0xf4, 0xff, 0xff},
		}),
		// specify that the button's text needs some padding for correct display
		widget.ButtonOpts.TextPadding(widget.Insets{
			Left:  30,
			Right: 30,
		}),
		widget.ButtonOpts.ClickedHandler(clickToRestart),
	)
	rootContainer.AddChild(button)
	rootContainer.AddChild(button2)
	textWidget = widget.NewText(textInfo)
	rootContainer.AddChild(textWidget)
	GUIhandler = &ebitenui.UI{Container: rootContainer}
	return GUIhandler
}

func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return GameWidth, GameHeight
}

func loadPNGImageFromEmbedded(name string) *ebiten.Image {
	pictNames, err := EmbeddedAssets.ReadDir("assets")
	if err != nil {
		log.Fatal("failed to read embedded dir ", pictNames, " ", err)
	}
	embeddedFile, err := EmbeddedAssets.Open("assets/" + name)
	if err != nil {
		log.Fatal("failed to load embedded image ", embeddedFile, err)
	}
	rawImage, err := png.Decode(embeddedFile)
	if err != nil {
		log.Fatal("failed to load embedded image ", name, err)
	}
	gameImage := ebiten.NewImageFromImage(rawImage)
	return gameImage
}

func processPlayerInput(theGame *Game) {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		theGame.player.dY = -PlayerSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		theGame.player.dY = PlayerSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyUp) || inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		theGame.player.dY = 0
	}
	theGame.player.yloc += theGame.player.dY
	if theGame.player.yloc <= 0 {
		theGame.player.dY = 0
		theGame.player.yloc = 0
	} else if theGame.player.yloc+theGame.player.pict.Bounds().Size().Y > GameHeight {
		theGame.player.dY = 0
		theGame.player.yloc = GameHeight - theGame.player.pict.Bounds().Size().Y
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		theGame.player.dX = -PlayerSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		theGame.player.dX = PlayerSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyArrowRight) || inpututil.IsKeyJustReleased(ebiten.KeyArrowLeft) {
		theGame.player.dX = 0
	}
	theGame.player.xloc += theGame.player.dX
	if theGame.player.xloc <= 0 {
		theGame.player.dX = 0
		theGame.player.xloc = 0
	} else if theGame.player.xloc+theGame.player.pict.Bounds().Size().X > GameWidth {
		theGame.player.dX = 0
		theGame.player.xloc = GameWidth - theGame.player.pict.Bounds().Size().X
	}
}

func collision(player, enemy Sprite) bool {
	goldWidth, goldHeight := enemy.pict.Size()
	playerWidth, playerHeight := player.pict.Size()
	if player.xloc < enemy.xloc+goldWidth &&
		player.xloc+playerWidth > enemy.xloc &&
		player.yloc < enemy.yloc+goldHeight &&
		player.yloc+playerHeight > enemy.yloc {
		return true
	}
	return false
}

func remove(slice []Sprite, s int) []Sprite {
	return append(slice[:s], slice[s+1:]...)
}

func clickToQuit(args *widget.ButtonClickedEventArgs) {
	os.Exit(0)
}

func clickToRestart(args *widget.ButtonClickedEventArgs) {
	g.enemy = g.fillEnemySlice(g.enemy)
	g.showUI = false
}

func (g *Game) fillEnemySlice(slice []Sprite) []Sprite {
	g.enemy = make([]Sprite, 10)
	for i := 0; i < len(g.enemy); i++ {
		g.enemy[i] = Sprite{
			pict: loadPNGImageFromEmbedded("lpc_goblin_preview.png"),
			xloc: rand.Intn(GameWidth - 50),
			yloc: rand.Intn(GameHeight - 50),
		}
	}
	return g.enemy
}

// make enemy slice call own function

// https://stackoverflow.com/questions/37334119/how-to-delete-an-element-from-a-slice-in-golang
// https://www.digitalocean.com/community/tutorials/understanding-init-in-go
// https://github.com/jsantore/FirstGameDemo/blob/master/GmeEngineDemo.go
