package webadvisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/jacobmichels/Course-Sense-Go/internal/types"
)

type WebAdvisor struct {
	client *http.Client
}

type WebAdvisorCourseListResponse struct {
	Courses []struct {
		MatchingSectionIds []string
		Id                 string
		SubjectCode        string
		Number             string
	}
}

type WebAdvisorSectionListRequest struct {
	CourseId   string   `json:"courseId"`
	SectionIds []string `json:"sectionIds"`
}

type WebAdvisorSectionListResponse struct {
	TermsAndSections []struct {
		Sections []struct {
			Section struct {
				Capacity  uint
				Available uint
				CourseId  string
				Id        string
				Number    string
				TermId    string
			}
		}
	}
	Course struct {
		Id string
	}
}

func NewWebAdvisor() (*WebAdvisor, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &WebAdvisor{
		client: &http.Client{
			Jar: jar,
		},
	}, nil
}

func (wa *WebAdvisor) GetSectionCapacity(ctx context.Context, section *types.CourseSection) (uint, error) {
	// How this works:
	// 1. Send a request to any page on webadvisor to get a RequestVerificationToken
	// 2. Send a request to the course search endpoint to get the course id and section ids
	// 3. Send a request to the section list endpoint to get the section info

	// I'm doing three requests for each notification request, that's dumb
	// i should only need to do 1-2 requests per, the token can be reused
	// maybe implement some kind of cache for the sections ids

	token, err := wa.getRequestVerificationToken(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get request verification token: %w", err)
	}

	courseId, sectionIds, err := wa.searchCourses(ctx, token, section)
	if err != nil {
		return 0, fmt.Errorf("failed to search courses: %w", err)
	}

	capacity, err := wa.getCapacity(ctx, section.Term, section.SectionCode, token, courseId, sectionIds)
	if err != nil {
		return 0, fmt.Errorf("failed to get capacity: %w", err)
	}

	return capacity, nil
}

func (wa *WebAdvisor) getRequestVerificationToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://colleague-ss.uoguelph.ca/Student/Courses", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	res, err := wa.client.Do(req)
	if err != nil {
		return "", err
	}

	token, err := extractToken(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to extract token: %w", err)
	}

	return token, nil
}

func (wa *WebAdvisor) searchCourses(ctx context.Context, token string, section *types.CourseSection) (string, []string, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"searchParameters":"{\"keyword\":null,\"terms\":[],\"requirement\":null,\"subrequirement\":null,\"courseIds\":null,\"sectionIds\":null,\"requirementText\":null,\"subrequirementText\":\"\",\"group\":null,\"startTime\":null,\"endTime\":null,\"openSections\":null,\"subjects\":[\"%s\"],\"academicLevels\":[],\"courseLevels\":[],\"synonyms\":[],\"courseTypes\":[],\"topicCodes\":[],\"days\":[],\"locations\":[],\"faculty\":[],\"onlineCategories\":null,\"keywordComponents\":[],\"startDate\":null,\"endDate\":null,\"startsAtTime\":null,\"endsByTime\":null,\"pageNumber\":1,\"sortOn\":\"None\",\"sortDirection\":\"Ascending\",\"subjectsBadge\":[],\"locationsBadge\":[],\"termFiltersBadge\":[],\"daysBadge\":[],\"facultyBadge\":[],\"academicLevelsBadge\":[],\"courseLevelsBadge\":[],\"courseTypesBadge\":[],\"topicCodesBadge\":[],\"onlineCategoriesBadge\":[],\"openSectionsBadge\":\"\",\"openAndWaitlistedSectionsBadge\":\"\",\"subRequirementText\":null,\"quantityPerPage\":500,\"openAndWaitlistedSections\":null,\"searchResultsView\":\"CatalogListing\"}"}`, section.Department))
	req, err := http.NewRequestWithContext(ctx, "POST", "https://colleague-ss.uoguelph.ca/Student/Courses/SearchAsync", data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json, charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("__RequestVerificationToken", token)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	res, err := wa.client.Do(req)
	if err != nil {
		return "", nil, err
	}

	var courseList WebAdvisorCourseListResponse
	err = json.NewDecoder(res.Body).Decode(&courseList)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode json: %w", err)
	}

	for _, course := range courseList.Courses {
		if course.SubjectCode == section.Department && course.Number == fmt.Sprintf("%d", section.CourseCode) {
			return course.Id, course.MatchingSectionIds, nil
		}
	}

	return "", nil, fmt.Errorf("course not found")
}

func (wa *WebAdvisor) getCapacity(ctx context.Context, term, targetSectionId, token, courseId string, sectionIds []string) (uint, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"courseId":"%s","sectionIds":%s}`, courseId, sectionIds))
	req, err := http.NewRequestWithContext(ctx, "POST", "https://colleague-ss.uoguelph.ca/Student/Courses/SectionsAsync", data)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json, charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("__RequestVerificationToken", token)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	res, err := wa.client.Do(req)
	if err != nil {
		return 0, err
	}

	var sectionList WebAdvisorSectionListResponse
	err = json.NewDecoder(res.Body).Decode(&sectionList)
	if err != nil {
		return 0, fmt.Errorf("failed to decode json: %w", err)
	}

	for _, termAndSection := range sectionList.TermsAndSections {
		for _, webAdvisorSection := range termAndSection.Sections {
			if webAdvisorSection.Section.Number == targetSectionId && webAdvisorSection.Section.TermId == term {
				return webAdvisorSection.Section.Available, nil
			}
		}
	}

	return 0, fmt.Errorf("section not found")
}

func extractToken(body io.Reader) (string, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}
	bodyString := string(bodyBytes)
	lines := strings.Split(bodyString, "\n")

	var token string
	for _, line := range lines {
		if strings.Contains(line, `<input name="__RequestVerificationToken" type="hidden" value="`) {
			token = line[strings.Index(line, `value="`)+7 : strings.Index(line, `" />`)]
		}
	}

	if token == "" {
		return "", fmt.Errorf("token not found")
	}

	return token, nil
}
