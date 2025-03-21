package controllers

import (
	"net/http"

	"kgent-api/api/services"

	"github.com/gin-gonic/gin"
)

type ResourceCtl struct {
	resourceService *services.ResourceService
}

func NewResourceCtl(service *services.ResourceService) *ResourceCtl {
	return &ResourceCtl{resourceService: service}
}

func (r *ResourceCtl) List() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		if resource == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource parameter is required"})
			return
		}

		ns := c.DefaultQuery("ns", "default")

		resourceList, err := r.resourceService.ListResource(c.Request.Context(), resource, ns)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": resourceList})
	}
}

func (r *ResourceCtl) Delete() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		if resource == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource parameter is required"})
			return
		}

		ns := c.DefaultQuery("ns", "default")
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name parameter is required"})
			return
		}

		err := r.resourceService.DeleteResource(c.Request.Context(), resource, ns, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": "resource deleted successfully"})
	}
}

func (r *ResourceCtl) Create() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		if resource == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource parameter is required"})
			return
		}

		type ResourceParam struct {
			Yaml string `json:"yaml" binding:"required"`
		}

		var param ResourceParam
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := r.resourceService.CreateResource(c.Request.Context(), resource, param.Yaml)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": "resource created successfully"})
	}
}

func (r *ResourceCtl) GetGVR() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Query("resource")
		if resource == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "resource parameter is required"})
			return
		}

		gvr, err := r.resourceService.GetGVR(resource)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": *gvr})
	}
}
