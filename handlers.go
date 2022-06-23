package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func checkboxGetHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "checkbox.html", nil)
}

func checkboxPostHandler(c *gin.Context) {
	var checkboxForm checkboxForm
	if err := c.Bind(&checkboxForm); err != nil {
		c.String(http.StatusBadRequest, err.Error())
	} else {
		c.JSON(http.StatusOK, gin.H{"color": checkboxForm.Colors})
	}
}

func profileHandler(c *gin.Context) {
	var profileForm profileForm
	if err := c.ShouldBind(&profileForm); err != nil {
		// if err := c.ShouldBindWith(&profileForm, binding.Form); err != nil {
		c.String(http.StatusBadRequest, "bind error", err.Error())
	} else {
		err := c.SaveUploadedFile(profileForm.Avatar, "/Users/kopever/Develop/temp/gin-demo/"+profileForm.Name)
		if err != nil {
			c.String(http.StatusInternalServerError, "save error", err.Error())
		} else {
			c.String(http.StatusOK, "ok")
		}
	}
}
