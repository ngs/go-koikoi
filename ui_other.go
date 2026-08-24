//go:build !js

package main

import "github.com/ngs/gocui"

// registerWebResize はブラウザ以外では何もしない。
func registerWebResize(*gocui.Gui) {}
