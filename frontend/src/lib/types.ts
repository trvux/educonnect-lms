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

export type Material = {
  id: number;
  lesson_id: number;
  file_name: string;
  file_path: string;
};

export type EnrolledStudent = {
  StudentID: number;
  FullName: string;
  Email: string;
};
