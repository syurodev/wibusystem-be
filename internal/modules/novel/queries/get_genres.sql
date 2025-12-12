-- GetGenres: Lấy danh sách genre IDs của novel
SELECT genre_id FROM catalog.novel_genres WHERE novel_id = $1
