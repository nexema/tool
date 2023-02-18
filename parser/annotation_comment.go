package parser

type annotationOrCommentEntry struct {
	line   int
	values []annotationOrComment
}

type annotationOrComment struct {
	annotation *AnnotationStmt
	comment    *CommentStmt
}

func annotationOrCommentEntryComparator(a, b annotationOrCommentEntry) bool {
	return a.line < b.line
}

func newAnnotationOrCommentArray() *[]annotationOrComment {
	return new([]annotationOrComment)
}
