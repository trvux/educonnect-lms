// Lệnh seed bơm dữ liệu THẬT (không hardcode nội dung giả) vào EduConnect
// LMS bằng cách gọi thẳng HTTP API đã có sẵn (giống 1 client bình thường),
// không insert SQL trực tiếp. Nguồn nội dung là giáo trình môn "Quản lý dự
// án phần mềm" thật của người dùng (đã export sẵn ra .md tại SEED_SOURCE_DIR).
//
// Chạy: (cần backend đang chạy ở SEED_API_BASE_URL, mặc định localhost:8080)
//
//	go run ./cmd/seed
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	baseURL   = envOr("SEED_API_BASE_URL", "http://localhost:8080/api")
	sourceDir = envOr("SEED_SOURCE_DIR", "/Users/tranvux/Downloads/qlda/markdown_files")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---- tài khoản seed --------------------------------------------------
// Giảng viên dùng tên thật của chủ đồ án (chủ sở hữu khóa học). Admin/học
// viên chỉ là tài khoản tiện ích để tạo dữ liệu demo (duyệt khóa học, nộp
// bài) — không đại diện cho người thật nào.
const (
	teacherEmail    = "tranvu@vlu.edu.vn"
	teacherPassword = "TranVu@2026"
	teacherName     = "Trần Vũ"

	adminEmail    = "admin.seed@vlu.edu.vn"
	adminPassword = "AdminSeed@2026"
	adminName     = "Admin Seed"

	studentEmail    = "student.seed@vlu.edu.vn"
	studentPassword = "StudentSeed@2026"
	studentName     = "Sinh Viên Seed"
)

// ---- nội dung khóa học (map file thật -> chương/bài học) --------------
type topicSpec struct {
	ChapterTitle string
	LessonTitle  string
	SourceFile   string
	ExtraFiles   []string
}

var topics = []topicSpec{
	{
		ChapterTitle: "00. Giới thiệu môn học",
		LessonTitle:  "Giới thiệu môn học Quản lý dự án phần mềm",
		SourceFile:   "00_Giới thiệu môn học QLDA PM.md",
		ExtraFiles: []string{
			"71ITSE41503_Quản lý dự án phần mềm_253_71ITSE41503_01_TIEUL_1_GUISV.md",
			"BM-006_71ITSE41003_QLDAPM_252_71ITSE41003_01_02_03_TIEUL_1.md",
		},
	},
	{
		ChapterTitle: "01. Tổng quan về QLDA Phần mềm",
		LessonTitle:  "Tổng quan về Quản lý dự án Phần mềm",
		SourceFile:   "01_Tổng quan về QLDA Phần mềm.md",
	},
	{
		ChapterTitle: "02. Khởi động dự án",
		LessonTitle:  "Khởi động dự án",
		SourceFile:   "Slide-02_Khởi động dự án.md",
	},
	{
		ChapterTitle: "03. Quy trình và framework phát triển phần mềm",
		LessonTitle:  "Quy trình và framework phát triển phần mềm",
		SourceFile:   "Slide-03_Quy trình và framework phát triển phần mềm.md",
	},
	{
		ChapterTitle: "04. Lập kế hoạch cho dự án",
		LessonTitle:  "Lập kế hoạch cho dự án",
		SourceFile:   "Slide-04_Lập kế hoạch cho DA.md",
		ExtraFiles:   []string{"gantt chart template.md"},
	},
	{
		ChapterTitle: "05. Ước lượng, chất lượng và truyền thông",
		LessonTitle:  "Ước lượng, chất lượng và truyền thông trong dự án",
		SourceFile:   "Slide-05_Lập kế hoạch cho DA_Ước lượng-Chất lượng - Truyền thông.md",
		ExtraFiles: []string{
			"05_Ước lượng cho dự án phần mềm.md",
			"Story point.md",
			"Velocity.md",
			"Individual Estimate.md",
			"PM - Assignment05.md",
		},
	},
	{
		ChapterTitle: "06. Lập kế hoạch quản lý rủi ro",
		LessonTitle:  "Lập kế hoạch quản lý rủi ro",
		SourceFile:   "Slide-06_Lập KH_QL Rủi ro.md",
	},
	{
		ChapterTitle: "07. Thực thi dự án",
		LessonTitle:  "Thực thi dự án",
		SourceFile:   "Slide-07_Thực thi Dự án.md",
	},
	{
		ChapterTitle: "08. Giám sát và kiểm soát dự án",
		LessonTitle:  "Giám sát và kiểm soát dự án (Monitoring and Controlling)",
		SourceFile:   "Chương 8 - Monitoring and Controlling.md",
	},
	{
		ChapterTitle: "09. Git Flow và tùy biến Scrum",
		LessonTitle:  "Git Flow và tùy biến Scrum",
		SourceFile:   "09_Git Flow and Customize Scrum.md",
		ExtraFiles:   []string{"scrum process template.md"},
	},
	{
		ChapterTitle: "10. Đóng dự án",
		LessonTitle:  "Đóng dự án",
		SourceFile:   "Slide-10_Đóng dự án.md",
	},
}

// quizByLesson: câu hỏi trắc nghiệm do seed tự soạn bám sát nội dung thật
// của từng bài học (không có sẵn ngân hàng câu hỏi trong tài liệu gốc).
type quizQuestion struct {
	Content      string
	Options      []string
	CorrectIndex int
}

var quizByLesson = map[string][]quizQuestion{
	"Khởi động dự án": {
		{
			Content:      "Tài liệu nào xác định mục tiêu, phạm vi và trao quyền chính thức cho Project Manager khi khởi động dự án?",
			Options:      []string{"Project Charter", "Sprint Backlog", "Burndown Chart", "Retrospective Report"},
			CorrectIndex: 0,
		},
	},
	"Ước lượng, chất lượng và truyền thông trong dự án": {
		{
			Content:      "Story Point dùng để đo lường điều gì?",
			Options:      []string{"Thời gian tuyệt đối (giờ/ngày) để hoàn thành", "Khối lượng công việc tương đối của User Story", "Số lượng thành viên cần thiết", "Chi phí bằng tiền của dự án"},
			CorrectIndex: 1,
		},
		{
			Content:      "Velocity của nhóm Scrum được xác định như thế nào?",
			Options:      []string{"Tổng Story Point ước lượng ban đầu của Product Backlog", "Số Story Point nhóm hoàn thành thực tế trong 1 Sprint", "Số giờ làm việc trung bình mỗi ngày", "Số lượng bug phát sinh mỗi Sprint"},
			CorrectIndex: 1,
		},
	},
	"Lập kế hoạch quản lý rủi ro": {
		{
			Content:      "Trong ma trận đánh giá rủi ro (Probability x Impact), rủi ro rơi vào vùng đỏ nên được xử lý như thế nào?",
			Options:      []string{"Bỏ qua vì hiếm khi xảy ra", "Ưu tiên xử lý/theo dõi sát vì mức độ nghiêm trọng cao", "Chỉ ghi nhận, không cần hành động", "Chuyển giao toàn bộ cho khách hàng"},
			CorrectIndex: 1,
		},
	},
	"Giám sát và kiểm soát dự án (Monitoring and Controlling)": {
		{
			Content:      "Burndown Chart trong Sprint thể hiện điều gì?",
			Options:      []string{"Số lượng thành viên còn lại trong nhóm", "Khối lượng công việc (Story Point) còn lại theo thời gian", "Tổng chi phí đã chi tiêu", "Số lượng rủi ro đã xảy ra"},
			CorrectIndex: 1,
		},
	},
	"Git Flow và tùy biến Scrum": {
		{
			Content:      "Trong Git Flow, nhánh nào dùng để tích hợp các nhánh feature đã hoàn thành trước khi chuẩn bị release?",
			Options:      []string{"main", "develop", "hotfix", "feature"},
			CorrectIndex: 1,
		},
	},
}

// ---- HTTP client tối giản ----------------------------------------------

type apiClient struct {
	http  *http.Client
	token string
}

func (c *apiClient) postJSON(path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *apiClient) uploadFile(path string, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return c.do(req, nil)
}

func (c *apiClient) do(req *http.Request, out any) error {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s -> %d: %s", req.Method, req.URL.Path, resp.StatusCode, string(body))
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("decode response lỗi: %w (body=%s)", err, body)
		}
	}
	return nil
}

// registerOrLogin đăng ký tài khoản mới; nếu email đã tồn tại (409) thì
// đăng nhập bằng tài khoản cũ — giúp script chạy lại an toàn.
func registerOrLogin(c *apiClient, email, password, fullName, role string) (token string, userID uint, err error) {
	var user struct {
		ID uint `json:"id"`
	}
	regErr := c.postJSON("/auth/register", map[string]string{
		"email": email, "password": password, "full_name": fullName, "role": role,
	}, &user)
	if regErr != nil {
		fmt.Printf("  (tài khoản %s có thể đã tồn tại, thử đăng nhập)\n", email)
	}

	var login struct {
		Token string `json:"token"`
	}
	if err := c.postJSON("/auth/login", map[string]string{"email": email, "password": password}, &login); err != nil {
		return "", 0, fmt.Errorf("đăng nhập %s thất bại: %w", email, err)
	}
	return login.Token, user.ID, nil
}

func mustFile(name string) string {
	p := filepath.Join(sourceDir, name)
	if _, err := os.Stat(p); err != nil {
		panic(fmt.Sprintf("không tìm thấy file nguồn %q: %v", p, err))
	}
	return p
}

func main() {
	c := &apiClient{http: &http.Client{Timeout: 30 * time.Second}}

	fmt.Println("== 1. Tạo/đăng nhập tài khoản ==")
	teacherToken, teacherID, err := registerOrLogin(c, teacherEmail, teacherPassword, teacherName, "teacher")
	must(err)
	fmt.Printf("Giảng viên: %s (id=%d)\n", teacherEmail, teacherID)

	adminToken, _, err := registerOrLogin(c, adminEmail, adminPassword, adminName, "admin")
	must(err)
	fmt.Printf("Admin seed: %s\n", adminEmail)

	studentToken, _, err := registerOrLogin(c, studentEmail, studentPassword, studentName, "student")
	must(err)
	fmt.Printf("Học viên seed: %s\n", studentEmail)

	fmt.Println("\n== 2. Tạo khóa học ==")
	teacherClient := &apiClient{http: c.http, token: teacherToken}
	adminClient := &apiClient{http: c.http, token: adminToken}
	studentClient := &apiClient{http: c.http, token: studentToken}

	var course struct {
		ID uint `json:"id"`
	}
	must(teacherClient.postJSON("/courses", map[string]string{
		"title": "Quản lý dự án phần mềm",
		"description": "Môn học giải thích ý nghĩa và tầm quan trọng của việc quản lý thành công 1 dự án phần mềm: " +
			"lựa chọn quy trình phát triển phù hợp, lập kế hoạch và giám sát tiến độ, ước lượng khối lượng công việc/chi phí, " +
			"quản lý thay đổi và rủi ro phát sinh trong dự án. Nội dung khóa học được số hóa từ giáo trình thật, học kỳ 3 năm học 2024-2025.",
	}, &course))
	fmt.Printf("Course id=%d\n", course.ID)

	must(teacherClient.postJSON(fmt.Sprintf("/courses/%d/submit", course.ID), map[string]string{}, nil))
	must(adminClient.postJSON(fmt.Sprintf("/admin/courses/%d/approve", course.ID), map[string]string{}, nil))
	fmt.Println("Đã gửi duyệt + duyệt khóa học (approved)")

	fmt.Println("\n== 3. Tạo chương/bài học + upload tài liệu thật ==")
	// lessonIDByTitle giúp gắn Assignment/Quiz vào đúng bài học ở bước sau.
	lessonIDByTitle := map[string]uint{}

	for _, t := range topics {
		var chapter struct {
			ID uint `json:"id"`
		}
		must(teacherClient.postJSON(fmt.Sprintf("/courses/%d/chapters", course.ID), map[string]string{"title": t.ChapterTitle}, &chapter))

		var lesson struct {
			ID uint `json:"id"`
		}
		must(teacherClient.postJSON(fmt.Sprintf("/chapters/%d/lessons", chapter.ID), map[string]string{"title": t.LessonTitle}, &lesson))
		lessonIDByTitle[t.LessonTitle] = lesson.ID

		must(teacherClient.uploadFile(fmt.Sprintf("/lessons/%d/materials", lesson.ID), mustFile(t.SourceFile)))
		for _, extra := range t.ExtraFiles {
			must(teacherClient.uploadFile(fmt.Sprintf("/lessons/%d/materials", lesson.ID), mustFile(extra)))
		}
		fmt.Printf("  [%s] %s (%d tài liệu)\n", t.ChapterTitle, t.LessonTitle, 1+len(t.ExtraFiles))
	}

	fmt.Println("\n== 4. Tạo bài tập (essay từ tài liệu ước lượng thật + quiz tự soạn) ==")
	estimationLessonID := lessonIDByTitle["Ước lượng, chất lượng và truyền thông trong dự án"]
	dueAt := time.Now().AddDate(0, 0, 14).UTC().Format(time.RFC3339)

	var essay1 struct {
		ID uint `json:"id"`
	}
	must(teacherClient.postJSON(fmt.Sprintf("/lessons/%d/assignments", estimationLessonID), map[string]any{
		"title": "Bài tập ước lượng cá nhân (Individual Estimate)",
		"description": "Dựa trên tài liệu \"Individual Estimate\" đính kèm (WBS mẫu theo các giai đoạn Initiate/Planning/Implement), " +
			"mỗi sinh viên tự ước lượng lại thời gian (giờ) cho từng task và trình bày cách ước lượng của mình.",
		"kind":   "essay",
		"due_at": dueAt,
	}, &essay1))

	var essay2 struct {
		ID uint `json:"id"`
	}
	must(teacherClient.postJSON(fmt.Sprintf("/lessons/%d/assignments", estimationLessonID), map[string]any{
		"title": "Bài tập ước lượng nhóm (Assignment 05)",
		"description": "Dựa trên tài liệu \"PM - Assignment05\" đính kèm, thảo luận nhóm và cập nhật cột Final estimate, " +
			"giải thích chênh lệch giữa các lần Change so với Estimate ban đầu.",
		"kind":   "essay",
		"due_at": dueAt,
	}, &essay2))
	fmt.Printf("  essay: id=%d, id=%d\n", essay1.ID, essay2.ID)

	quizIDByLesson := map[string]uint{}
	for lessonTitle, questions := range quizByLesson {
		lessonID, ok := lessonIDByTitle[lessonTitle]
		if !ok {
			continue
		}
		qs := make([]map[string]any, 0, len(questions))
		for _, q := range questions {
			qs = append(qs, map[string]any{
				"content":       q.Content,
				"options":       q.Options,
				"correct_index": q.CorrectIndex,
			})
		}
		var quiz struct {
			ID uint `json:"id"`
		}
		must(teacherClient.postJSON(fmt.Sprintf("/lessons/%d/assignments", lessonID), map[string]any{
			"title":       "Trắc nghiệm: " + lessonTitle,
			"description": "Câu hỏi ôn tập nhanh nội dung bài học (do giảng viên biên soạn).",
			"kind":        "quiz",
			"questions":   qs,
		}, &quiz))
		quizIDByLesson[lessonTitle] = quiz.ID
		fmt.Printf("  quiz [%s]: id=%d (%d câu)\n", lessonTitle, quiz.ID, len(questions))
	}

	fmt.Println("\n== 5. Học viên seed đăng ký + làm bài (để có dữ liệu demo chấm điểm/tiến độ) ==")
	must(studentClient.postJSON(fmt.Sprintf("/courses/%d/enroll", course.ID), map[string]string{}, nil))

	// Nộp thử 2 bài trắc nghiệm (tự động chấm điểm ngay).
	for lessonTitle := range quizByLesson {
		quizID, ok := quizIDByLesson[lessonTitle]
		if !ok {
			continue
		}
		answers := make([]int, len(quizByLesson[lessonTitle]))
		for i, q := range quizByLesson[lessonTitle] {
			answers[i] = q.CorrectIndex // học viên seed làm đúng hết để demo điểm 10
		}
		if err := studentClient.postJSON(fmt.Sprintf("/assignments/%d/submit", quizID), map[string]any{"answers": answers}, nil); err != nil {
			fmt.Printf("  (bỏ qua nộp quiz %q: %v)\n", lessonTitle, err)
		}
	}

	// Nộp bài tự luận rồi để giảng viên chấm — demo luồng US5.3.
	var submission struct {
		ID uint `json:"id"`
	}
	must(studentClient.postJSON(fmt.Sprintf("/assignments/%d/submit", essay1.ID), map[string]any{
		"content": "Em ước lượng lại các task theo kinh nghiệm nhóm: phần Initiate khoảng 14h, Planning khoảng 3h, " +
			"Implement tăng dần theo độ phức tạp thực tế phát sinh trong quá trình phát triển.",
	}, &submission))
	must(teacherClient.postJSON(fmt.Sprintf("/submissions/%d/grade", submission.ID), map[string]any{
		"score":    8.5,
		"feedback": "Trình bày rõ ràng, cần giải thích thêm lý do chênh lệch giữa các lần ước lượng.",
	}, nil))
	fmt.Println("  Đã nộp + chấm 1 bài tự luận demo")

	fmt.Println("\n== 6. Diễn đàn + thông báo ==")
	var question struct {
		ID uint `json:"id"`
	}
	must(studentClient.postJSON(fmt.Sprintf("/courses/%d/forum-posts", course.ID), map[string]any{
		"content": "Thầy/cô cho em hỏi sự khác nhau giữa Story Point và Velocity ạ? Em đọc tài liệu vẫn hơi rối.",
	}, &question))
	must(teacherClient.postJSON(fmt.Sprintf("/courses/%d/forum-posts", course.ID), map[string]any{
		"content":   "Story Point ước lượng độ lớn tương đối của công việc, còn Velocity là số Story Point nhóm thực sự hoàn thành mỗi Sprint — em xem lại tài liệu Story point.md/Velocity.md đính kèm ở bài Ước lượng nhé.",
		"parent_id": question.ID,
	}, nil))

	must(teacherClient.postJSON(fmt.Sprintf("/courses/%d/notifications", course.ID), map[string]string{
		"title":   "Đã cập nhật tài liệu bài giảng",
		"message": "Lớp mình đã có đầy đủ slide + tài liệu tham khảo cho tất cả các bài học, mời các bạn xem trong từng bài học của khóa học.",
	}, nil))
	fmt.Println("  Đã tạo 1 câu hỏi + 1 trả lời diễn đàn, 1 thông báo")

	fmt.Println("\n✅ Seed xong. Đăng nhập bằng:")
	fmt.Printf("   Giảng viên: %s / %s\n", teacherEmail, teacherPassword)
	fmt.Printf("   Học viên seed: %s / %s\n", studentEmail, studentPassword)
	fmt.Printf("   Admin seed: %s / %s\n", adminEmail, adminPassword)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
