package controllers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"kgent-api/api/services"

	"github.com/gin-gonic/gin"
)

type PodLogEventCtl struct {
	podLogEventService *services.PodLogEventService
}

func NewPodLogEventCtl(service *services.PodLogEventService) *PodLogEventCtl {
	return &PodLogEventCtl{podLogEventService: service}
}

func (p *PodLogEventCtl) GetLog() func(c *gin.Context) {
	return func(c *gin.Context) {
		ns := c.DefaultQuery("ns", "default")
		podname := c.DefaultQuery("podname", "")
		container := c.DefaultQuery("container", "")

		tailLineStr := c.DefaultQuery("tailLine", "100")
		tailLine, err := strconv.ParseInt(tailLineStr, 10, 64)
		if err != nil {
			tailLine = 100
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		req, err := p.podLogEventService.GetLogs(ctx, ns, podname, container, tailLine)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		rc, err := req.Stream(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		defer rc.Close()

		logData, err := io.ReadAll(rc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": string(logData)})
	}
}

func (p *PodLogEventCtl) GetEvent() func(c *gin.Context) {
	return func(c *gin.Context) {
		ns := c.DefaultQuery("ns", "default")
		podname := c.DefaultQuery("podname", "")

		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()

		e, err := p.podLogEventService.GetEvents(ctx, ns, podname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": e})
	}
}
