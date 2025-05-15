package main

import (
	"fmt"

	"lenslockd/models"
)

func main() {
	gs := models.GalleryService{}
	fmt.Println(gs.Images(2))
}