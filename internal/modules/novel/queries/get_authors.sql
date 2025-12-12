-- GetAuthors: Lấy danh sách author IDs của novel
SELECT author_id FROM catalog.novel_authors WHERE novel_id = $1
