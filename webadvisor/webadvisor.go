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

	coursesense "github.com/jacobmichels/Course-Sense-Go"
)

var _ coursesense.SectionService = WebAdvisorSectionService{}

type WebAdvisorSectionService struct {
	http *http.Client
}

func NewWebAdvisorSectionService() (WebAdvisorSectionService, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return WebAdvisorSectionService{}, fmt.Errorf("failed to instantiate cookiejar")
	}

	client := &http.Client{
		Jar: jar,
	}

	return WebAdvisorSectionService{client}, nil
}

func (w WebAdvisorSectionService) Exists(ctx context.Context, section coursesense.Section) (bool, error) {
	token, err := w.getRequestVerificationToken(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get request verification token: %w", err)
	}

	courseID, sectionIDs, err := w.searchCourses(ctx, token, section)
	if err != nil {
		return false, fmt.Errorf("failed to search for course: %w", err)
	}

	webAdvisorSections, err := w.listSections(ctx, token, courseID, sectionIDs)
	if err != nil {
		return false, fmt.Errorf("failed to list sections: %w", err)
	}

	for _, webAdvisorSection := range webAdvisorSections {
		if webAdvisorSection.Section.Number == section.Code && webAdvisorSection.Section.TermId == section.Term {
			return true, nil
		}
	}

	return false, nil
}

func (w WebAdvisorSectionService) GetAvailableSeats(ctx context.Context, section coursesense.Section) (uint, error) {
	token, err := w.getRequestVerificationToken(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get request verification token: %w", err)
	}

	courseID, sectionIDs, err := w.searchCourses(ctx, token, section)
	if err != nil {
		return 0, fmt.Errorf("failed to search for course: %w", err)
	}

	webAdvisorSections, err := w.listSections(ctx, token, courseID, sectionIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to list sections: %w", err)
	}

	for _, webAdvisorSection := range webAdvisorSections {
		if webAdvisorSection.Section.Number == section.Code && webAdvisorSection.Section.TermId == section.Term {
			return webAdvisorSection.Section.Available, nil
		}
	}

	return 0, fmt.Errorf("section not found")
}

func (w WebAdvisorSectionService) getRequestVerificationToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://colleague-ss.uoguelph.ca/Student/Courses", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	res, err := w.http.Do(req)
	if err != nil {
		return "", err
	}

	token, err := extractToken(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to extract token: %w", err)
	}

	return token, nil
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

type CourseSearchResponse struct {
	Courses []struct {
		MatchingSectionIds []string
		Id                 string
		SubjectCode        string
		Number             string
	}
}

// Returns the course ID, and section IDs
func (w WebAdvisorSectionService) searchCourses(ctx context.Context, token string, section coursesense.Section) (string, []string, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"searchParameters":"{\"keyword\":null,\"terms\":[],\"requirement\":null,\"subrequirement\":null,\"courseIds\":null,\"sectionIds\":null,\"requirementText\":null,\"subrequirementText\":\"\",\"group\":null,\"startTime\":null,\"endTime\":null,\"openSections\":null,\"subjects\":[\"%s\"],\"academicLevels\":[],\"courseLevels\":[],\"synonyms\":[],\"courseTypes\":[],\"topicCodes\":[],\"days\":[],\"locations\":[],\"faculty\":[],\"onlineCategories\":null,\"keywordComponents\":[],\"startDate\":null,\"endDate\":null,\"startsAtTime\":null,\"endsByTime\":null,\"pageNumber\":1,\"sortOn\":\"None\",\"sortDirection\":\"Ascending\",\"subjectsBadge\":[],\"locationsBadge\":[],\"termFiltersBadge\":[],\"daysBadge\":[],\"facultyBadge\":[],\"academicLevelsBadge\":[],\"courseLevelsBadge\":[],\"courseTypesBadge\":[],\"topicCodesBadge\":[],\"onlineCategoriesBadge\":[],\"openSectionsBadge\":\"\",\"openAndWaitlistedSectionsBadge\":\"\",\"subRequirementText\":null,\"quantityPerPage\":500,\"openAndWaitlistedSections\":null,\"searchResultsView\":\"CatalogListing\"}"}`, section.Course.Department))
	req, err := http.NewRequestWithContext(ctx, "POST", "https://colleague-ss.uoguelph.ca/Student/Courses/SearchAsync", data)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json, charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("__RequestVerificationToken", token)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	res, err := w.http.Do(req)
	if err != nil {
		return "", nil, err
	}

	var courseList CourseSearchResponse
	err = json.NewDecoder(res.Body).Decode(&courseList)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode json: %w", err)
	}

	for _, course := range courseList.Courses {
		if course.SubjectCode == section.Course.Department && course.Number == fmt.Sprintf("%d", section.Course.Code) {
			return course.Id, course.MatchingSectionIds, nil
		}
	}

	return "", nil, fmt.Errorf("%s*%d*%s*%s not found", section.Course.Department, section.Course.Code, section.Term, section.Code)
}

type SectionListResponse struct {
	TermsAndSections []struct {
		Sections []WebAdvisorSection
	}
	Course struct {
		Id string
	}
}

type WebAdvisorSection struct {
	Section struct {
		Capacity  uint
		Available uint
		CourseId  string
		Id        string
		Number    string
		TermId    string
	}
}

func (w WebAdvisorSectionService) listSections(ctx context.Context, token, courseId string, sectionIds []string) ([]WebAdvisorSection, error) {
	data := bytes.NewBufferString(fmt.Sprintf(`{"courseId":"%s","sectionIds":%s}`+"\n", courseId, "[\""+strings.Join(sectionIds, "\",\"")+"\"]"))
	req, err := http.NewRequestWithContext(ctx, "POST", "https://colleague-ss.uoguelph.ca/Student/Courses/SectionsAsync", data)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json, charset=utf-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("__RequestVerificationToken", token)
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	res, err := w.http.Do(req)
	if err != nil {
		return nil, err
	}

	var sectionList SectionListResponse
	err = json.NewDecoder(res.Body).Decode(&sectionList)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}

	var results []WebAdvisorSection
	for _, termAndSection := range sectionList.TermsAndSections {
		results = append(results, termAndSection.Sections...)
	}

	return results, nil
}
