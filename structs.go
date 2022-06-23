package main

import "mime/multipart"

type checkboxForm struct {
	Colors []string `form:"colors[]"`
}

type profileForm struct {
	Name   string                `form:"name" binding:"required"`
	Avatar *multipart.FileHeader `form:"avatar" binding:"required"`
}
