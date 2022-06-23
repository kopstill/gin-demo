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
