package main

import (
	"crypto/md5"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

type GridPoint struct {
	value byte
	index int
}

type Point struct {
	x, y int
}

type DrawingPoint struct {
	topLeft     Point
	bottomRight Point
}

type User struct {
	email     string
	ip        string
	publicKey string
}
type Avatar struct {
	hash       [16]byte
	grid       []byte
	gridPoints []GridPoint
	pixelMap   []DrawingPoint
	color      color.RGBA
	img        image.Image
}

type Apply func(Avatar) Avatar

func pipe(avatar Avatar, funcs ...Apply) Avatar {
	for _, applyer := range funcs {
		avatar = applyer(avatar)
	}
	return avatar
}
func createAvatar(user User) Avatar {
	email := []byte(user.email)
	ip := []byte(user.ip)
	publicKey := []byte(user.publicKey)
	info := append(append(email, ip...), publicKey...)
	checkSum := md5.Sum(info)
	return Avatar{
		hash: checkSum,
	}
}

func pickColor(avatar Avatar) Avatar {
	avatar.color = color.RGBA{avatar.hash[0], avatar.hash[1], avatar.hash[2], 255}
	return avatar
}

func buildGrid(avatar Avatar) Avatar {
	grid := []byte{}

	for i := 0; i < len(avatar.hash) && i+3 <= len(avatar.hash)-1; i += 3 {

		chunk := make([]byte, 5)
		copy(chunk, avatar.hash[i:i+3])
		chunk[3] = chunk[1]
		chunk[4] = chunk[0]
		grid = append(grid, chunk...)
	}
	avatar.grid = grid
	return avatar
}

func filterOddSquares(avatar Avatar) Avatar {
	grid := []GridPoint{}
	for i, code := range avatar.grid {
		if code%2 == 0 {
			point := GridPoint{
				value: code,
				index: i,
			}
			grid = append(grid, point)
		}
	}
	avatar.gridPoints = grid
	return avatar
}

func buildPixelMap(avatar Avatar) Avatar {
	drawingPoints := []DrawingPoint{}

	pixelFunc := func(p GridPoint) DrawingPoint {

		horizontal := (p.index % 5) * 50

		vertical := (p.index / 5) * 50

		topLeft := Point{horizontal, vertical}

		bottomRight := Point{horizontal + 50, vertical + 50}

		return DrawingPoint{
			topLeft,
			bottomRight,
		}
	}

	for _, gridPoint := range avatar.gridPoints {
		drawingPoints = append(drawingPoints, pixelFunc(gridPoint))
	}
	avatar.pixelMap = drawingPoints
	return avatar
}

func fill(img image.NRGBA, start Point, end Point, color color.Color) {
	for i := start.x; i <= end.x; i++ {
		for j := start.y; j <= end.y; j++ {
			img.Set(i, j, color)
		}
	}
}

func drawAvatar(avatar Avatar) Avatar {
	img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{250, 250}})
	for _, pixel := range avatar.pixelMap {
		fill(image.NRGBA(*img), pixel.topLeft, pixel.bottomRight, avatar.color)
	}
	avatar.img = img
	return avatar
}

func saveImage(avatar Avatar) {
	f, err := os.Create("outimage.png")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	err = png.Encode(f, avatar.img)
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	user := User{
		email:     "aliahmedismail37@gmail.com",
		ip:        "127.0.0.1",
		publicKey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0",
	}

	avi := createAvatar(user)

	avi = pipe(avi, pickColor, buildGrid, filterOddSquares, buildPixelMap, drawAvatar)

	saveImage(avi)
}
