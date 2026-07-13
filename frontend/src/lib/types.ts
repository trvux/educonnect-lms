export type CourseStatus = "draft" | "pending_review" | "approved";

export type Course = {
  id: number;
  title: string;
  description: string;
  status: CourseStatus;
  teacher_id: number;
};

export type Chapter = {
  id: number;
  course_id: number;
  title: string;
  position: number;
};

export type Lesson = {
  id: number;
  chapter_id: number;
  title: string;
  position: number;
};

// US4.10 — trạng thái hoàn thành/khóa của 1 bài học, trả về từ
// GET /courses/:id/lesson-progress (theo góc nhìn người dùng hiện tại).
export type LessonProgressEntry = {
  lesson_id: number;
  completed: boolean;
  locked: boolean;
};

// US4.9/US4.10 — bài học trong outline có kèm completed/locked (chỉ có ý
// nghĩa với học viên; giảng viên/admin luôn completed=false, locked=false).
export type OutlineLesson = Lesson & { completed: boolean; locked: boolean };

// US4.9 — chương kèm danh sách bài học lồng sẵn, dùng cho sidebar course player.
export type CourseOutlineChapter = Chapter & { lessons: OutlineLesson[] };

export type MaterialFileType = "pdf" | "doc" | "excel" | "ppt" | "video" | "archive";

export type Material = {
  id: number;
  lesson_id: number;
  file_name: string;
  file_path: string;
  file_type: MaterialFileType;
  // US4.5 — chỉ có khi file_type === "video"; token ngắn hạn để phát trực
  // tiếp qua /materials/:id/stream (xem lib/api/materials.ts).
  stream_token?: string;
};

export type EnrolledStudent = {
  StudentID: number;
  FullName: string;
  Email: string;
};

// Epic 5 — US5.1/US5.2/US5.3
export type AssignmentKind = "essay" | "quiz";

export type Question = {
  content: string;
  options: string[];
  // correct_index chỉ có khi người xem là giảng viên/quản trị viên đã đăng
  // nhập (backend ẩn field này với học viên/khách — xem OptionalAuth).
  correct_index?: number;
};

export type Assignment = {
  id: number;
  lesson_id: number;
  title: string;
  description: string;
  kind: AssignmentKind;
  questions: Question[];
  due_at?: string;
  // US5.4 — chỉ có ở bài trắc nghiệm; có giá trị nghĩa là học viên phải làm
  // xong trong X phút kể từ lúc bắt đầu (POST /assignments/:id/start-attempt).
  time_limit_minutes?: number;
};

// US5.4 — trả về từ POST /assignments/:id/start-attempt.
export type QuizAttempt = {
  started_at: string;
};

export type Submission = {
  id: number;
  assignment_id: number;
  student_id: number;
  content: string;
  answers: number[];
  score?: number;
  feedback?: string;
  graded: boolean;
};

export type GradebookEntry = {
  student_id: number;
  student_name: string;
  assignment_id: number;
  assignment_title: string;
  score?: number;
};

// Epic 6 — US6.1/US6.2
export type ForumPost = {
  id: number;
  course_id: number;
  author_id: number;
  author_name?: string;
  parent_id?: number;
  content: string;
  created_at: string;
};

export type Notification = {
  id: number;
  course_id: number;
  title: string;
  message: string;
  read: boolean;
  created_at: string;
};

// Epic 7 — US7.1/US7.2
export type CourseProgress = {
  course_id: number;
  course_title: string;
  total_assignments: number;
  submitted: number;
  percent_complete: number;
};

export type CourseStats = {
  course_id: number;
  course_title: string;
  enrolled_students: number;
  total_assignments: number;
  average_completion: number;
};

// Sprint 3 — US1.4/1.5/1.6/1.7/1.8
export type UserProfile = {
  id: number;
  email: string;
  full_name: string;
  role: "student" | "teacher" | "admin";
  phone?: string;
  student_code?: string;
  avatar_path?: string;
  email_verified: boolean;
};

export type RoleUpgradeStatus = "pending" | "approved" | "rejected";

export type RoleUpgradeRequest = {
  id: number;
  user_id: number;
  reason: string;
  status: RoleUpgradeStatus;
  reviewed_by?: number;
  created_at: string;
  reviewed_at?: string;
};
