// main.go
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Post struct {
	PostTitle               string
	Format                  string
	ReleaseTime            string
	Tags                    string
	RadarChart             string
	Views                   int
	Likes                   int
	Collects               int
	Comments               int
	GrowFollowers          int
	Shared                 int
	TrafficSource          string
	FemalePercentage       string
	Age2534Percentage      string
	Age1824Percentage      string
	OverseasPercentage     string
	InterestDistribution   string
	GrowthStrategy         string
}


func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	// Open and read CSV file
	file, err := os.Open("creator_data.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var posts []Post

	// Skip header row
	_, err = reader.Read()
	if err != nil {
		fmt.Println("Error reading header:", err)
		return
	}

	// Read data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading record:", err)
			continue
		}

		// Convert string values to integers where needed
		views, _ := strconv.Atoi(record[5])      // Views
		likes, _ := strconv.Atoi(record[6])      // Likes
		collects, _ := strconv.Atoi(record[7])   // Collects
		comments, _ := strconv.Atoi(record[8])   // Comments
		growFollowers, _ := strconv.Atoi(record[9]) // GrowFollowers
		shared, _ := strconv.Atoi(record[10])    // Shared

		post := Post{
			PostTitle:             record[0],
			Format:               record[1],
			ReleaseTime:          record[2],
			Tags:                 record[3],
			RadarChart:           record[4],
			Views:                views,
			Likes:                likes,
			Collects:             collects,
			Comments:             comments,
			GrowFollowers:        growFollowers,
			Shared:               shared,
			TrafficSource:        record[11],
			FemalePercentage:     record[12],
			Age2534Percentage:    record[13],
			Age1824Percentage:    record[14],
			OverseasPercentage:   record[15],
			InterestDistribution: record[16],
			GrowthStrategy:       record[17],
		}
		posts = append(posts, post)
	}

	// Update sorting criteria to use Views + Likes + Comments + Collects
	sort.Slice(posts, func(i, j int) bool {
		engagementI := posts[i].Views + posts[i].Likes + posts[i].Comments + posts[i].Collects
		engagementJ := posts[j].Views + posts[j].Likes + posts[j].Comments + posts[j].Collects
		return engagementI > engagementJ
	})

	// Get top 5 posts
	topPosts := posts
	if len(posts) > 5 {
		topPosts = posts[:5]
	}

	// Add debug logging
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        fmt.Println("Warning: OPENAI_API_KEY is empty")
        return
    }

	// Prepare analysis prompt
	var analysisText strings.Builder
	analysisText.WriteString("Perform post analysis of the following top 5 high engagement posts:\n\n")
	
	for i, post := range topPosts {
		analysisText.WriteString(fmt.Sprintf("Post #%d:\n", i+1))
		analysisText.WriteString(fmt.Sprintf("Title: %s\n", post.PostTitle))
		analysisText.WriteString(fmt.Sprintf("Format: %s\n", post.Format))
		analysisText.WriteString(fmt.Sprintf("Release Time: %s\n", post.ReleaseTime))
		analysisText.WriteString(fmt.Sprintf("Tags: %s\n", post.Tags))
		analysisText.WriteString(fmt.Sprintf("Views: %d\n", post.Views))
		analysisText.WriteString(fmt.Sprintf("Likes: %d\n", post.Likes))
		analysisText.WriteString(fmt.Sprintf("Comments: %d\n", post.Comments))
		analysisText.WriteString(fmt.Sprintf("Collects: %d\n", post.Collects))
		analysisText.WriteString(fmt.Sprintf("Growth in Followers: %d\n", post.GrowFollowers))
		analysisText.WriteString(fmt.Sprintf("Shares: %d\n", post.Shared))
		analysisText.WriteString(fmt.Sprintf("Traffic Source: %s\n", post.TrafficSource))
		analysisText.WriteString(fmt.Sprintf("Female Percentage: %s\n", post.FemalePercentage))
		analysisText.WriteString(fmt.Sprintf("Age 25-34 Percentage: %s\n", post.Age2534Percentage))
		analysisText.WriteString(fmt.Sprintf("Age 18-24 Percentage: %s\n", post.Age1824Percentage))
		analysisText.WriteString(fmt.Sprintf("Overseas Percentage: %s\n", post.OverseasPercentage))
		analysisText.WriteString(fmt.Sprintf("Interest Distribution: %s\n", post.InterestDistribution))
		analysisText.WriteString(fmt.Sprintf("Growth Strategy: %s\n\n", post.GrowthStrategy))
	}

	// Replace the OpenAI client code with direct HTTP request
	url := "https://api.openai.com/v1/chat/completions"
	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": analysisText.String(),
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	// Add headers including Bearer authentication
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read and parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	// Check if the response status is successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error from API: %v\n", result)
		return
	}

	// Safely access the response data with checks
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		fmt.Println("No choices in response")
		return
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid choice format")
		return
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		fmt.Println("Invalid message format")
		return
	}

	content, ok := message["content"].(string)
	if !ok {
		fmt.Println("Invalid content format")
		return
	}

	analysis := content

	// Write analysis to file
	err = os.WriteFile("creator_analysis.txt", []byte(analysis), 0644)
	if err != nil {
		fmt.Printf("Error writing analysis to file: %v\n", err)
		return
	}

	fmt.Println("Analysis of top 5 posts completed and saved to creator_analysis.txt")
}

