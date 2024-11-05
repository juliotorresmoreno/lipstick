package manager

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/lipstick/helper"
	"github.com/juliotorresmoreno/lipstick/server/auth"
	"github.com/juliotorresmoreno/lipstick/server/config"
)

type router struct {
	manager *Manager
}

func configureRouter(manager *Manager) {
	router := &router{manager: manager}
	r := gin.New()

	r.GET("/health", router.health)
	r.GET("/ws", router.upgrade)
	r.GET("/ws/:uuid", router.request)

	r.GET("/domains", router.getDomains)
	r.GET("/domains/:domain_name", router.getDomain)
	r.POST("/domains", router.addDomain)
	r.PATCH("/domains/:domain_name", router.updateDomain)
	r.DELETE("/domains/:domain_name", router.deleteDomain)

	manager.engine = r
}

func isAuthorized(c *gin.Context) bool {
	conf, err := config.GetConfig()
	if err != nil {
		log.Println("Unable to get config", err)
		return false
	}
	authorization := c.Request.Header.Get("Authorization")
	return authorization == conf.Keyword
}

func (r *router) getDomains(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	domains, err := r.manager.authManager.GetDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get domains"})
		return
	}

	c.JSON(http.StatusOK, domains)
}

func (r *router) getDomain(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	domain_name := c.Param("domain_name")
	domain, err := r.manager.authManager.GetDomain(domain_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get domain"})
		return
	}

	c.JSON(http.StatusOK, domain)
}

func (r *router) addDomain(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	domain := &auth.Domain{}
	if err := c.BindJSON(domain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.manager.authManager.AddDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *router) updateDomain(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	domain := &auth.Domain{}
	if err := c.BindJSON(domain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain_name := c.Param("domain_name")
	record, err := r.manager.authManager.GetDomain(domain_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get domain"})
		return
	}

	domain.ID = record.ID
	if err := r.manager.authManager.UpdateDomain(domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *router) deleteDomain(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	domain_name := c.Param("domain_name")
	record, err := r.manager.authManager.GetDomain(domain_name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get domain"})
		return
	}

	if err := r.manager.authManager.DelDomain(record.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete domain"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *router) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (r *router) upgrade(c *gin.Context) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Unable to upgrade connection", err)
		return
	}

	domainName, err := helper.GetDomainName(wsConn.NetConn())
	if err != nil {
		log.Println("Unable to get domain name", err)
		wsConn.Close()
		return
	}

	domain, err := r.manager.authManager.GetDomain(domainName)
	if err != nil {
		log.Println("Unable to get domain", err)
		wsConn.Close()
		return
	}

	r.manager.registerWebsocketConn <- &websocketConn{
		Domain: domain.Name,
		Conn:   wsConn,
	}
}

func (r *router) request(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Unable to upgrade connection", err)
		return
	}

	uuid, ok := c.Params.Get("uuid")
	if !ok {
		return
	}
	r.manager.request <- &request{uuid: uuid, conn: conn}
}
