package parser

type annotationOrComment struct {
	annotation *AnnotationStmt
	comment    *CommentStmt
}

type annotationOrCommentEntry struct {
	line   int
	values *[]annotationOrComment
}

func annotationOrCommentEntryComparator(a, b annotationOrCommentEntry) bool {
	return a.line < b.line
}

func unwrapAnnotationsOrComments(original []annotationOrComment, annotations *[]AnnotationStmt, comments *[]CommentStmt) {
	if original == nil {
		return
	}

	for _, elem := range original {
		if elem.annotation != nil {
			*annotations = append(*annotations, *elem.annotation)
		} else if elem.comment != nil {
			*comments = append(*comments, *elem.comment)
		}
	}
}
