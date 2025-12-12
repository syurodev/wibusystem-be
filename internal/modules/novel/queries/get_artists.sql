-- GetArtists: Lấy danh sách artist IDs của novel
SELECT artist_id FROM catalog.novel_artists WHERE novel_id = $1
