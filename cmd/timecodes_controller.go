package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	timecodeParser "timecodes/cmd/timecode_parser"
)

type TimecodeJSON struct {
	ID          uint   `json:"id"`
	Description string `json:"description"`
	LikesCount  int    `json:"likesCount"`
	LikedByMe   bool   `json:"likedByMe"`
	Seconds     int    `json:"seconds"`
	VideoID     string `json:"videoId"`
}

// GET /timecodes
func handleGetTimecodes(c *Container, w http.ResponseWriter, r *http.Request) {
	currentUser := getCurrentUser(r)
	videoID := mux.Vars(r)["videoId"]

	timecodes, err := c.TimecodeRepository.FindByVideoId(videoID)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		if len(*timecodes) == 0 {
			go func() {
				parseDescriptionAndCreateAnnotations(c, videoID)
				parseCommentsAndCreateAnnotations(c, videoID)
			}()
		}

		timecodeJSONCollection := make([]*TimecodeJSON, 0)
		for _, timecode := range *timecodes {
			timecodeJSONCollection = append(timecodeJSONCollection, serializeTimecode(timecode, currentUser))
		}

		json.NewEncoder(w).Encode(timecodeJSONCollection)
	}
}

// POST /timecodes
func handleCreateTimecode(c *Container, w http.ResponseWriter, r *http.Request) {
	currentUser := getCurrentUser(r)
	timecode := &Timecode{}

	reqBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(reqBody, timecode)
	_, err := c.TimecodeRepository.Create(timecode)

	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(serializeTimecode(timecode, currentUser))
	}
}

func serializeTimecode(timecode *Timecode, currentUser *User) (timecodeJSON *TimecodeJSON) {
	var likedByMe bool
	if currentUser != nil {
		likedByMe = getLikedByMe(timecode.Likes, currentUser.ID)
	}

	return &TimecodeJSON{
		ID:          timecode.ID,
		Description: timecode.Description,
		LikesCount:  len(timecode.Likes),
		LikedByMe:   likedByMe,
		Seconds:     timecode.Seconds,
		VideoID:     timecode.VideoID,
	}
}

func getLikedByMe(likes []TimecodeLike, userID uint) bool {
	for _, like := range likes {
		if like.UserID == userID {
			return true
		}
	}

	return false
}

func parseDescriptionAndCreateAnnotations(c *Container, videoID string) {
	description := c.YoutubeAPI.FetchVideoDescription(videoID)
	parsedCodes := timecodeParser.Parse(description)

	_, err := c.TimecodeRepository.CreateFromParsedCodes(parsedCodes, videoID)
	if err != nil {
		log.Println(err)
	}
}

func parseCommentsAndCreateAnnotations(c *Container, videoID string) {
	var parsedCodes []timecodeParser.ParsedTimeCode

	comments, err := c.YoutubeAPI.FetchVideoComments(videoID)
	if err != nil {
		log.Println(err)

		return
	}

	for _, comment := range comments {
		timeCodes := timecodeParser.Parse(comment.Snippet.TopLevelComment.Snippet.TextOriginal)

		parsedCodes = append(parsedCodes, timeCodes...)
	}

	_, err = c.TimecodeRepository.CreateFromParsedCodes(parsedCodes, videoID)
	if err != nil {
		log.Println(err)
	}
}
