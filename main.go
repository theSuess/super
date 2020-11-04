package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	jwt "github.com/cristalhq/jwt/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	fileBucketName = []byte("fileMeta")
)

type SubmissionMeta struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"time"`
	Files     map[string]string `json:"files"`
}

func main() {

	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <datadir> <password>\n", os.Args[0])
		os.Exit(1)
	}

	datadir := os.Args[1]
	password := os.Args[2]

	jwtSigner, err := jwt.NewSignerHS(jwt.HS256, []byte(password))
	if err != nil {
		log.Err(err).Msg("could not construct jwtSigner")
	}
	jwtBuilder := jwt.NewBuilder(jwtSigner)

	jwtVerifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(password))
	if err != nil {
		log.Err(err).Msg("could not construct jwtVerifier")
	}

	r := gin.Default()

	withAuth := func(c *gin.Context) {
		token := c.GetHeader("X-Auth-Token")
		if token != password {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid or missing password",
			})
			return
		}
		c.Next()
	}

	r.GET("/download/:submission/:file", func(c *gin.Context) {
		token, err := jwt.ParseAndVerifyString(c.Query("token"), jwtVerifier)
		if err != nil {
			log.Err(err).Msg("bad jwt token")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "bad jwt token",
			})
			return
		}
		var claims jwt.StandardClaims
		json.Unmarshal(token.RawClaims(), &claims)
		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "download token expired",
			})
			return
		}
		fp := path.Join(datadir, c.Param("submission"), c.Param("file"))
		log.Info().Str("path", fp).Msg("sending file")
		c.FileAttachment(fp, c.Query("name"))
	})
	r.GET("/list", withAuth, func(c *gin.Context) {
		submissions := []SubmissionMeta{}
		err := filepath.Walk(datadir, func(path string, info os.FileInfo, err error) error {
			if info.Name() != "meta.json" {
				return nil
			}
			raw, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			meta := SubmissionMeta{}
			if err := json.Unmarshal(raw, &meta); err != nil {
				return err
			}
			submissions = append(submissions, meta)
			return nil
		})
		if err != nil {
			log.Err(err).Msg("unable to walk directory tree")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "unable to walk directory tree",
			})
			return
		}
		downloadToken, _ := jwtBuilder.Build(&jwt.StandardClaims{
			ExpiresAt: &jwt.NumericDate{
				Time: time.Now().Add(time.Minute * 15),
			},
		})
		c.Header("X-Download-Token", downloadToken.String())
		c.JSON(http.StatusOK, submissions)
	})

	r.POST("/upload", func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload[]"]
		submissionID, _ := uuid.NewRandom()

		submission := SubmissionMeta{
			ID:        submissionID.String(),
			Name:      c.Request.Form.Get("name"),
			Timestamp: time.Now(),
			Files:     map[string]string{},
		}

		if submission.Name == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "name must not be empty",
			})
			return
		}

		log.Info().Str("submitter", submission.Name).Msg("new submission")

		err := os.Mkdir(path.Join(datadir, submissionID.String()), 0755)
		if err != nil {
			log.Err(err).Msg("submission directory could not be created")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "submission directory could not be created",
			})
			return
		}

		for _, file := range files {
			fid, _ := uuid.NewUUID()
			err := c.SaveUploadedFile(file, path.Join(datadir, submissionID.String(), fid.String()))
			if err != nil {
				log.Err(err).Msg("could not save file")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "could not save file",
				})
				return
			}
			submission.Files[fid.String()] = file.Filename
			log.Info().Str("filename", file.Filename).Msg("file uploaded")
		}

		raw, err := json.Marshal(submission)
		if err != nil {
			log.Err(err).Msg("could not marshal metadata")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "could not marshal metadata",
			})
			return
		}

		err = ioutil.WriteFile(path.Join(datadir, submissionID.String(), "meta.json"), raw, 0644)
		if err != nil {
			log.Err(err).Msg("could not save metadata")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "could not save metadata",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "uploaded files",
			"count":   len(files),
			"id":      submissionID.String(),
			"time":    submission.Timestamp,
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Err(err).Msg("listening for requests")
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}
