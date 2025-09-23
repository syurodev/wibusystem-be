Thiết kế Hệ thống CSDL cho Nền tảng Quản lý Anime/Manga/Novel

***Phân hệ Anime – Cấu trúc Dữ liệu Chính***

Thực thể cốt lõi: Phân hệ Anime quản lý các bộ anime và nội dung liên quan. Mỗi Anime (tác phẩm) là thực thể cha đại diện cho một series anime, chứa thông tin tổng quan (tên, mô tả, ảnh bìa, trạng thái phát hành, v.v.). Thuộc tính mở rộng của anime bao gồm mùa phát sóng (Xuân/Hạ/Thu/Đông) và năm phát sóng – cho biết anime bắt đầu chiếu vào mùa nào, năm nào. Một anime có thể phân chia thành nhiều Season (mùa) – mỗi season là tập hợp các tập phim (Season 1, Season 2, …). Mỗi Episode (tập phim) là đơn vị nội dung chính chứa liên kết video stream và phụ đề tương ứng. Tập phim có số thứ tự trong mùa, tiêu đề tập và độ dài; hệ thống có thể lưu URL video và hỗ trợ nhiều phụ đề. Dưới đây là các bảng chính:

	•	anime – Danh sách anime (series). Mỗi bản ghi có id (UUID) làm khóa chính, trạng thái (đang chiếu, đã hoàn thành, tạm ngưng) dạng enum content_status, mùa và năm phát sóng đầu tiên, ảnh bìa…
	•	anime_season – Các mùa anime, liên kết với bảng anime. Có khóa chính id (UUID), khóa ngoại anime_id trỏ đến anime, số mùa (season_number) để sắp xếp (Season 1, 2,…), tiêu đề mùa (nếu có). Có thể lưu thêm thông tin kinh doanh: giá mua cả mùa, giá thuê và thời hạn thuê (ngày) cho mùa đó.
	•	anime_episode – Các tập phim, liên kết theo mùa. Chứa id (UUID) khóa chính, season_id khóa ngoại (tập thuộc mùa nào), số tập trong mùa (episode_number) và tiêu đề tập phim. Thuộc tính quan trọng: cờ is_public (tập này có miễn phí hay cần trả phí), và price_coins (giá mua lẻ tập, dùng đơn vị tiền ảo nền tảng). Có thể lưu độ dài (duration) và URL video của tập.
	•	episode_subtitle – (Tùy chọn) Bảng quản lý phụ đề cho từng tập. Mỗi bản ghi lưu episode_id tập phim, language_code mã ngôn ngữ của phụ đề (vd: “en”, “vi”), và đường dẫn hoặc URL file phụ đề. Bảng này cho phép nhiều phụ đề ngôn ngữ cho một tập, hỗ trợ tính năng đóng góp phụ đề cộng đồng.

Quan hệ & Ràng buộc: Một anime có nhiều mùa (1 – N), một mùa có nhiều tập phim (1 – N). Các ràng buộc UNIQUE đảm bảo tính duy nhất: (anime_id, season_number) trong bảng season, (season_id, episode_number) trong bảng episode – không có hai mùa trùng số thứ tự trong cùng một anime, không có hai tập trùng số trong cùng một mùa. Khi xóa anime, các mùa và tập liên quan được xóa theo (ON DELETE CASCADE).

Hệ thống cũng quản lý Nhân vật và Diễn viên lồng tiếng (Seiyuu) cho anime: Mỗi nhân vật anime có thể do một seiyuu (diễn viên) lồng tiếng, và một diễn viên có thể lồng nhiều nhân vật. Để thể hiện, bảng trung gian anime_character liên kết anime – character – creator (voice actor). Mỗi bản ghi anime_character chứa khóa ngoại anime_id, character_id (nhân vật) và voice_actor_id (một creator có vai trò diễn viên). Quan hệ này cho phép: một anime có nhiều nhân vật (1 – N), một nhân vật xuất hiện trong nhiều anime (N – N), và mỗi bản ghi có tối đa một diễn viên (hoặc null nếu chưa có thông tin). Ta đặt UNIQUE(anime_id, character_id) để tránh trùng nhân vật trong cùng anime. Bảng character (xem phần Hệ thống dùng chung) lưu thông tin nhân vật (tên, mô tả, ảnh) dùng chung cho toàn nền tảng. Tương tự, creator dùng chung cho diễn viên. Mỗi bản ghi anime_character kết nối 1 nhân vật với 1 anime; nếu có voice_actor_id thì cho biết diễn viên nào lồng tiếng nhân vật đó trong anime. (Lưu ý: cột voice_actor_id dùng ON DELETE SET NULL – nếu diễn viên bị xóa khỏi hệ thống, quan hệ vẫn giữ nhân vật nhưng bỏ trống diễn viên).

Hệ thống Anime cũng tích hợp studio sản xuất và các nhà sáng tạo khác thông qua bảng liên kết anime_creator. Bảng này có anime_id, creator_id và role (enum creator_role). Ví dụ: một anime có thể liên kết đến một creator với role = STUDIO (hãng phim sản xuất). Nếu anime có nguyên tác (vd. chuyển thể từ novel), có thể thêm creator role AUTHOR cho tác giả nguyên tác, v.v. Ràng buộc UNIQUE(anime_id, creator_id, role) để một creator chỉ đóng một vai trò duy nhất trên một anime (nhưng một anime có thể có nhiều creator khác nhau và một creator có thể tham gia nhiều anime).

Thể loại (genre): Anime dùng chung hệ thống thể loại. Mỗi anime có thể thuộc nhiều thể loại (vd. Action, Fantasy…). Bảng liên kết anime_genre chứa anime_id và genre_id, liên kết nhiều-nhiều (N:N) giữa anime và bảng genre dùng chung. UNIQUE(anime_id, genre_id) đảm bảo không lặp thể loại.

Đa ngôn ngữ: Tiêu đề và mô tả của anime hỗ trợ nhiều ngôn ngữ. Bảng anime_translation chứa các bản dịch tiêu đề/mô tả cho từng anime. Mỗi bản ghi có anime_id, mã language_code (vd. “en”, “vi”), title và description bằng ngôn ngữ đó, và cờ is_primary để đánh dấu tiêu đề chính thức ở ngôn ngữ đó. Ví dụ, một anime có title gốc tiếng Nhật và title tiếng Anh; cả hai được lưu với language_code khác nhau, và cờ is_primary giúp xác định “tên chính” cho mỗi ngôn ngữ locale. Ràng buộc đảm bảo mỗi anime chỉ có một title chính per ngôn ngữ: dùng UNIQUE(anime_id, language_code) kèm điều kiện is_primary=true (partial index). Cách thiết kế tách bảng dịch thế này giúp giữ mô hình chuẩn hóa và dễ mở rộng ngôn ngữ mới mà không phải thêm cột.

***Phân hệ Manga – Cấu trúc Dữ liệu Chính***

Phân hệ Manga kế thừa nhiều ý tưởng từ Novel (cùng là nội dung đọc) nhưng tập trung vào truyện tranh hình ảnh. Mỗi Manga (tác phẩm truyện tranh) là thực thể cha đại diện cho một bộ truyện, có thuộc tính tương tự Novel: tên, tác giả, họa sĩ, mô tả, ảnh bìa, trạng thái phát hành (đang xuất bản, đã hoàn thành, v.v.). Bảng manga lưu các thông tin này (id UUID, status, cover_image,…). Một manga chứa nhiều Volume (tập truyện), mỗi volume nhóm các chương với nhau. Volume được thiết kế linh hoạt: đối với manga không xuất bản theo tập (ví dụ webtoon), hệ thống sẽ tự động tạo một “Volume Mặc định” để chứa các chương. Bảng manga_volume có khóa chính UUID, manga_id (FK), số volume (volume_number) và tiêu đề tập (nếu có), ảnh bìa và mô tả ngắn cho tập. Ràng buộc UNIQUE(manga_id, volume_number) để tránh trùng số tập trong một truyện. Volume_number có thể tự động là 1 cho “volume mặc định” nếu bộ truyện không phân tập chính thức.

Chapter (Chương): Mỗi manga có nhiều Chapter – đơn vị nội dung người dùng đọc theo chương. Bảng manga_chapter khóa chính UUID, volume_id FK (chương thuộc tập nào), số thứ tự chương trong volume, tiêu đề (nếu có), và cờ is_public (chương miễn phí hoặc trả phí). Mỗi chương thuộc đúng một volume và chứa danh sách các Page (trang truyện). Ràng buộc UNIQUE(volume_id, chapter_number) đảm bảo không trùng số chương trong cùng một tập. Bảng manga_page lưu các trang ảnh: chapter_id (FK), số trang và URL hình ảnh của trang. UNIQUE(chapter_id, page_number) để duy trì thứ tự trang duy nhất. Khi xóa chapter, các page con sẽ xoá theo (ON DELETE CASCADE).

Nhân vật & Tác giả: Tương tự anime, Manga liên kết với character (nhân vật) qua bảng manga_character (UUID, manga_id, character_id). Một nhân vật có thể xuất hiện ở nhiều manga (ví dụ cùng vũ trụ truyện) – do có cơ sở dữ liệu nhân vật trung tâm, ta có thể liên kết chéo, tạo trải nghiệm khám phá liền mạch giữa các phiên bản nội dung khác nhau. Bảng manga_character đảm bảo mỗi nhân vật chỉ khai báo một lần trong một manga (UNIQUE constraint). Manga cũng liên kết với các creator (tác giả, họa sĩ) qua bảng manga_creator (manga_id, creator_id, role). Vai trò có thể là AUTHOR (tác giả nội dung) hoặc ARTIST (họa sĩ vẽ truyện). Một người có thể vừa là tác giả vừa là họa sĩ cho cùng manga – khi đó sẽ có hai bản ghi với vai trò khác nhau (vd. một tác giả tự vẽ truyện mình viết). UNIQUE(manga_id, creator_id, role) ngăn trùng vai trò.

Thể loại: Bảng manga_genre (manga_id, genre_id) liên kết manga với các thể loại chung. Một manga có nhiều thể loại và thể loại có thể áp dụng cho nhiều manga (N:N). Ràng buộc duy nhất tương tự anime_genre.

Đa ngôn ngữ: Bảng manga_translation tương tự anime_translation, lưu tiêu đề và mô tả theo các ngôn ngữ khác nhau cho mỗi truyện. Mỗi bản ghi có manga_id, mã ngôn ngữ, title, description, và cờ is_primary (tiêu đề chính ở ngôn ngữ đó). Cách tách bảng giúp lưu nhiều tên gọi thay thế (Alias) cho manga – ví dụ tên gốc tiếng Nhật, tên tiếng Anh, v.v. – mà vẫn chuẩn hóa dữ liệu. Mỗi ngôn ngữ tối đa một tên chính (đảm bảo bởi UNIQUE index có điều kiện).

Quan hệ: Một manga có nhiều volume (1 – N); volume có nhiều chương (1 – N); chương có nhiều trang (1 – N). Xóa manga sẽ xóa cascade volumes, xóa volume sẽ xóa cascade chapters, xóa chapter xóa cascade pages. Quan hệ nhân vật và creator tương tự anime: xóa manga sẽ xóa các liên kết manga_character, manga_creator (ON DELETE CASCADE). Xóa character hay creator sẽ xóa các liên kết tương ứng (cascade trên character_id ở manga_character, creator_id ở manga_creator).

***Phân hệ Novel – Cấu trúc Dữ liệu Chính***

Phân hệ Novel quản lý tiểu thuyết (light novel, web novel, v.v.) dưới dạng nội dung văn bản thuần. Cấu trúc tương tự Manga nhưng thay trang ảnh bằng đoạn văn bản. Mỗi Novel (tác phẩm truyện chữ) là thực thể cha đại diện một bộ truyện hoàn chỉnh, có thuộc tính: tên truyện, tác giả, họa sĩ minh họa bìa (nếu có), mô tả, ảnh bìa, trạng thái (đang tiến hành/hoàn thành/tạm ngưng). Bảng novel (UUID PK) lưu các thông tin này.

Một novel chứa nhiều Volume (tập) – mỗi tập là một phần lớn của truyện (tương tự các tập sách). Bảng novel_volume có id (UUID), novel_id (FK), số tập, tên tập (ví dụ “Tập 1”), ảnh bìa tập và mô tả ngắn tập. Ràng buộc UNIQUE(novel_id, volume_number) để không trùng số tập. Mỗi volume chứa nhiều Chapter (chương) – đơn vị nội dung nhỏ nhất mà người dùng đọc. Bảng novel_chapter có id (UUID), volume_id (FK), số chương trong tập, tiêu đề chương, nội dung văn bản đầy đủ (field text), và thời gian xuất bản (timestamp). Cờ is_public và price_coins cũng có trong chapter (chương này miễn phí hay phải mua, và giá mua lẻ chương). Ràng buộc UNIQUE(volume_id, chapter_number) đảm bảo không trùng chương trong cùng tập.

Quan hệ: Novel 1 – N Volume, Volume 1 – N Chapter. Xóa novel sẽ xóa các volume, xóa volume sẽ xóa các chapter liên quan (ON DELETE CASCADE).

Nhân vật & Tác giả: Novel cũng liên kết nhân vật (nếu có nhân vật cụ thể trong truyện, ví dụ tiểu thuyết có nhân vật chính/phụ) qua novel_character (novel_id, character_id). Điều này cho phép tra cứu chéo nhân vật giữa bản novel và bản anime/manga chuyển thể, nếu cùng nhân vật. Novel liên kết creator qua novel_creator: thường gồm tác giả (AUTHOR) và họa sĩ minh họa (ILLUSTRATOR). Một số trường hợp tác giả tự minh họa thì sẽ có hai bản ghi cho cùng một người (2 vai trò). Ràng buộc unique tương tự các bảng trên.

Thể loại: Bảng novel_genre (novel_id, genre_id) liên kết novel với các thể loại (N:N), tương tự anime_genre.

Đa ngôn ngữ: Bảng novel_translation lưu tiêu đề/mô tả theo ngôn ngữ cho novel, thiết kế như các bảng translation khác. Novel cũng có thể có nhiều tên gọi (tên gốc, tên tiếng Anh, v.v.), mỗi tên ứng với một ngôn ngữ hoặc bút danh khác, được quản lý trong bảng này.

Nhóm sáng tạo sở hữu: Novel có trường nhóm sáng tạo sở hữu trong đặc tả (tức nhóm tác giả đăng tải). Chi tiết này có thể được lưu bằng cách coi nhóm cũng là một creator (kiểu tổ chức) và liên kết qua novel_creator với role STUDIO hoặc một role riêng (ví dụ OWNER). Tuy nhiên, để giữ thiết kế tổng quát, ta không thêm bảng riêng cho nhóm mà sử dụng cơ chế creator hiện có (có thể đánh dấu một creator là nhóm thay vì cá nhân trong dữ liệu).

***Hệ thống Dùng chung – Nhân vật, Nhà sáng tạo, Thể loại, Liên kết nội dung***

Nền tảng có các bảng dùng chung phục vụ cả 3 phân hệ:
	•	character – Danh sách Nhân vật hư cấu xuất hiện trong nội dung. Mỗi nhân vật có id (UUID), tên, mô tả, ảnh đại diện. Bảng này tập trung thông tin nhân vật để có thể liên kết một nhân vật với nhiều phiên bản nội dung khác nhau (anime, manga, novel). Các bảng liên kết anime_character, manga_character, novel_character (đã mô tả ở trên) kết nối nhân vật với từng tác phẩm cụ thể.
	•	creator – Danh sách Nhà sáng tạo nội dung (có thể là cá nhân hoặc tổ chức). Mỗi bản ghi có id (UUID), tên, mô tả. Bảng này bao gồm tác giả, họa sĩ, studio, diễn viên lồng tiếng,… không phân biệt loại ngay tại bảng – vai trò cụ thể của họ được thể hiện khi liên kết với nội dung. Điều này linh hoạt vì một người có thể tham gia nhiều vai trò (vd. vừa viết novel vừa lồng tiếng anime). Vai trò được định nghĩa trong enum creator_role với các giá trị ví dụ: AUTHOR, ARTIST (họa sĩ truyện tranh), ILLUSTRATOR (họa sĩ minh họa novel), STUDIO (hãng phim), VOICE_ACTOR (diễn viên lồng tiếng). Bảng liên kết anime_creator, manga_creator, novel_creator sử dụng trường role này để gán đúng vai trò. Ví dụ: một studio XYZ sẽ có một bản ghi trong creator, liên kết với một anime qua anime_creator với role = STUDIO; một tác giả “John Doe” có bản ghi creator, liên kết với tiểu thuyết A qua novel_creator role = AUTHOR, và cũng có thể liên kết với manga B qua manga_creator role = AUTHOR nếu ông viết cả hai. Tương tự, một diễn viên lồng tiếng liên kết với anime qua anime_character (do cần gắn với nhân vật cụ thể). Mô hình này giúp tránh trùng lặp dữ liệu người (một người có nhiều vai trò vẫn chỉ một bản ghi).
	•	genre – Danh mục Thể loại nội dung dùng chung. Mỗi thể loại có id (UUID) và tên thể loại (ví dụ: Action, Romance,…). Bảng genre liên kết đến anime/manga/novel qua các bảng trung gian như đã nêu. (Có thể mở rộng bảng genre với mô tả thể loại, hoặc đa ngôn ngữ nếu cần, nhưng hiện tại chỉ cần tên ngắn).
	•	content_relation – Bảng liên kết giữa các nội dung khác loại. Do hệ thống có nhiều dạng tác phẩm (anime, manga, novel) và một tác phẩm có thể là chuyển thể của tác phẩm khác (ví dụ anime chuyển thể từ manga), hoặc phần tiếp theo, ngoại truyện… ta thiết kế bảng quan hệ chung để mô tả những kết nối này. Bảng content_relation có các cột: source_id, source_type, target_id, target_type, và relation_type. Trong đó source_type và target_type là enum content_type_enum với các giá trị 'ANIME', 'MANGA', 'NOVEL' để chỉ loại nội dung của mỗi đầu mối; source_id và target_id chứa UUID của hai nội dung liên quan. relation_type là enum mô tả quan hệ: ví dụ 'ADAPTATION' (chuyển thể), 'SEQUEL' (phần tiếp theo), 'SPINOFF' (ngoại truyện), 'RELATED' (liên quan khác). Ví dụ: nếu anime X chuyển thể từ novel Y, ta tạo một bản ghi content_relation với source_type = ANIME, source_id = (id anime X), target_type = NOVEL, target_id = (id novel Y), relation_type = ADAPTATION. Bảng này linh hoạt cho mọi kiểu liên kết chéo. Mỗi quan hệ là hai chiều (có thể lưu hai bản ghi cho hai hướng, hoặc một chiều và khi truy vấn thì lấy cả hai; tùy ý). Trong thiết kế CSDL, do source/target có thể trỏ đến ba bảng khác nhau, ta không thể dùng ràng buộc khóa ngoại trực tiếp cho content_relation. Thay vào đó, ta quản lý tính hợp lệ ở tầng ứng dụng (khi thêm bản ghi phải đảm bảo ID tồn tại trong bảng tương ứng với type). Để tối ưu truy vấn, có thể tạo index ghép trên (source_type, source_id) và (target_type, target_id).

***Hỗ trợ Đa ngôn ngữ cho Nội dung***

Như đã đề cập, hệ thống hỗ trợ đa ngôn ngữ cho tiêu đề và mô tả của Anime, Manga, Novel. Cách triển khai là tách bảng phụ để lưu các bản dịch thay vì đưa nhiều cột song ngữ vào bảng chính. Điều này tuân theo thực tiễn thiết kế CSDL đa ngôn ngữ: tạo bảng bản dịch con chứa khóa chính của bảng mẹ kèm mã ngôn ngữ và nội dung dịch. Mô hình này giúp dễ bổ sung ngôn ngữ mới mà không thay đổi cấu trúc bảng chính, đồng thời giữ dữ liệu không bị lặp (mỗi ngôn ngữ một dòng thay vì thêm cột hoặc trộn JSON).

Cụ thể, mỗi bảng nội dung có bảng dịch tương ứng: anime_translation, manga_translation, novel_translation với schema tương tự nhau. Mỗi bản dịch gồm: id UUID, *_id (khóa ngoại đến nội dung), language_code (ví dụ ‘en’, ‘vi’, ‘ja’ – có thể theo chuẩn locale), title, description, is_primary. Đối với các nội dung đã có nhiều tên gọi thay thế (Alias) trong cùng một ngôn ngữ, ta lưu thành nhiều bản ghi với cùng language_code nhưng tiêu đề khác. Một trong số đó có is_primary = TRUE để chỉ định tên chính thức cho ngôn ngữ đó. Ta đặt unique index có điều kiện đảm bảo (content_id, language_code) là duy nhất khi is_primary = TRUE – nghĩa là mỗi nội dung mỗi ngôn ngữ chỉ được đánh dấu một tiêu đề chính. Người dùng sẽ thấy tiêu đề theo ngôn ngữ ưa thích (nếu có), và vẫn có thể xem các tên thay thế khác.

***Mô hình Thanh toán & Truy cập Nội dung***

Hệ thống tích hợp nhiều mô hình kiếm tiền song song nhằm tối đa hóa doanh thu và phục vụ đa dạng người dùng. Các mô hình bao gồm:
	•	Đọc công khai (Public): Một phần nội dung cho phép đọc miễn phí (có thể kèm quảng cáo). Thường các chương đầu (hoặc tập đầu) được đánh dấu miễn phí (is_public = TRUE). Logic kiểm tra truy cập đầu tiên sẽ xem nếu nội dung là public thì cho phép đọc ngay không cần trả tiền.
	•	Mua lẻ nội dung (Pay-per-Content): Người dùng dùng tiền ảo (Coin) để mua quyền truy cập vĩnh viễn một phần nội dung cụ thể. Đối với novel/manga, có thể mua từng chương, từng tập, hoặc toàn bộ tác phẩm với giá ưu đãi gói. Đối với anime, có thể mua từng tập hoặc theo mùa. Sau khi mua, nội dung đó mở khóa vĩnh viễn cho người dùng (lưu vào bảng lịch sử mua).
Thiết kế CSDL: Bảng user_content_purchases ghi lại các lượt mua. Mỗi bản ghi gồm user_id (UUID người dùng từ service User – cũng dùng UUID để nhất quán), item_type (loại nội dung mua) và item_id (UUID của nội dung cụ thể), cùng timestamp mua. Ta dùng enum purchase_item_type để chỉ rõ loại: ví dụ 'NOVEL_CHAPTER', 'NOVEL_VOLUME', 'NOVEL_SERIES', 'MANGA_CHAPTER', 'MANGA_VOLUME', 'MANGA_SERIES', 'ANIME_EPISODE', 'ANIME_SEASON' (và có thể 'ANIME_SERIES' nếu cho phép mua trọn bộ anime). Nhờ đó một bảng chung có thể lưu mọi giao dịch mua lẻ (thay vì tách bảng cho từng loại). Do item_id có thể trỏ đến nhiều bảng (episode, chapter, volume…), ta không dùng khóa ngoại cứng ở đây mà đảm bảo bằng logic ứng dụng. Bảng có index theo user_id để dễ truy vấn nội dung user đã mua (phục vụ kiểm tra quyền truy cập).
	•	Thuê nội dung (Rental): Cho phép người dùng thuê truyện/tập với giá rẻ hơn nhưng chỉ có quyền đọc trong thời hạn nhất định (ví dụ thuê một volume trong 7 ngày). Novel/manga: thường chỉ cho thuê sau khi tác phẩm hoàn thành, và áp dụng ở cấp Volume hoặc cả bộ truyện (Novel). Anime: có thể thuê theo season hoặc cả series khi đã phát hành đầy đủ. Khi thuê, user trả một khoản nhỏ để mở khóa tạm thời nội dung; sau expiry_date nội dung lại bị khóa.
Thiết kế: Bảng user_content_rentals ghi các lượt thuê. Trường item_type (enum rental_item_type) giới hạn các loại cho thuê: 'NOVEL_VOLUME', 'NOVEL_SERIES', 'MANGA_VOLUME', 'MANGA_SERIES', 'ANIME_SEASON', 'ANIME_SERIES'. Trường item_id trỏ đến nội dung tương ứng. Lưu rent_date (ngày giờ thuê) và expiry_date (ngày giờ hết hạn). Khi kiểm tra quyền, hệ thống sẽ đối chiếu expiry_date với hiện tại. Index trên user_id giúp truy vấn thuê của user nhanh. (Chú ý: Không cho thuê lẻ từng chương hay tập lẻ để giảm phức tạp; chỉ cấp lớn hơn như volume/season).
	•	Thuê bao (Subscription): Cung cấp gói thuê bao theo tháng/năm (Premium/VIP) để hưởng lợi ích đặc biệt. Lưu ý: thuê bao không tự động mở khóa toàn bộ nội dung tính phí (không biến user thường thành user trả phí toàn phần), mà thay vào đó cung cấp đặc quyền: đọc trước chương mới (Early Access), giảm giá mua/thuê, loại bỏ quảng cáo, v.v. Đặc biệt, gói VIP còn có thể có quyền truy cập tính năng độc quyền (ví dụ voice chat trong Watch Party của anime).
Thiết kế: Thông tin thuê bao người dùng có thể nằm bên service người dùng, nhưng để tích hợp, ta có bảng user_subscriptions trong hệ thống này để minh họa. Bảng gồm user_id, loại tier (enum subscription_tier: ví dụ FREE, PREMIUM, VIP), ngày bắt đầu và ngày hết hạn. Mỗi người dùng tại một thời điểm chỉ có một thuê bao hiệu lực (có thể dùng unique(user_id) để giới hạn một sub active, hoặc lưu nhiều dòng lịch sử). Khi kiểm tra quyền truy cập, nếu user có thuê bao VIP và chương ở trạng thái Early Access (phát hành sớm), họ được phép đọc trước. Ngoài ra, logic ứng dụng áp dụng giảm giá hoặc tắt quảng cáo dựa trên tier khi tính phí mua/thuê.
	•	Vé đọc hàng ngày (Daily Pass): Một mô hình khuyến khích người dùng miễn phí quay lại thường xuyên. Với các truyện đã hoàn thành, mỗi ngày người dùng nhận được một vé miễn phí để mở khóa tạm thời một chương. Nếu muốn đọc tiếp trong ngày phải trả tiền hoặc chờ hôm sau. Điều này giúp tăng tương tác hàng ngày.
Thiết kế: Không bắt buộc triển khai đầy đủ ở CSDL (có thể quản lý bằng bộ đếm trong bộ nhớ), nhưng nếu cần, có thể: (1) Thêm cột daily_pass_used trong bảng user_content_rentals cho biết lượt thuê miễn phí qua vé (0/1), hoặc (2) bảng user_daily_pass_log lưu mỗi lần dùng vé (user, chapter, date). Ở đây ta không chi tiết bảng, nhưng logic kiểm tra quyền sẽ có bước: nếu user có vé và chương đủ điều kiện (truyện hoàn thành, chương chưa đọc) thì cho phép đọc và trừ vé. Trừ vé có thể nghĩa là ghi lại vào log rằng ngày đó user đã dùng. Do đây là tính năng tùy chọn, ta chú thích thiết kế hơn là triển khai.
	•	Ủng hộ (Donate): Người dùng có thể donate (ủng hộ tiền) trực tiếp cho Nhóm Sáng tạo như một hình thức động viên. Donate không mở khóa thêm nội dung, mà chỉ là giao dịch một chiều (có thể xuất hiện trên hồ sơ nhóm sáng tạo hoặc xếp hạng nhà hảo tâm).
Thiết kế: Bảng user_donations lưu các khoản donate: user_id, content_type (ANIME/MANGA/NOVEL – loại nội dung mà user muốn ủng hộ, thường là toàn bộ tác phẩm hoặc cho nhóm tác giả tác phẩm đó), content_id (id cụ thể, ví dụ id novel được ủng hộ), số tiền amount, và donation_date. Thông qua content_type và content_id, hệ thống biết số tiền gắn với tác phẩm nào, từ đó có thể tổng hợp cho nhóm sáng tạo tương ứng. Mỗi donate là một bản ghi (có thể cả những donate nhỏ lẻ hoặc mua chương tặng bạn bè – nhưng khía cạnh “tặng quà” này có thể xem như donate hoặc giao dịch tách biệt).

Tích hợp với service Người dùng: Tất cả các bảng giao dịch trên sử dụng user_id là UUID để tham chiếu người dùng từ service ngoài. Có thể thiết lập foreign key trỏ tới bảng user (nếu cùng cơ sở dữ liệu), hoặc chỉ định rõ trong comment rằng user_id tham chiếu hệ thống User. Việc dùng UUID làm khóa chính cho mọi bảng (bao gồm user_id ngoại lai) giúp dễ dàng tích hợp, đảm bảo tính duy nhất trên hệ thống phân tán.

Kiểm tra quyền truy cập nội dung: Với các bảng trên, quy trình kiểm tra truy cập một chương/tập diễn ra tuần tự: (1) Nếu nội dung là public -> cho phép ngay; (2) nếu user là admin/mod -> cho phép; (3) nếu user có thuê bao và nội dung trong giai đoạn Early Access -> cho phép; (4) nếu user đã mua nội dung đó -> cho phép (tra bảng user_content_purchases); (5) nếu user đang thuê nội dung đó và chưa hết hạn -> cho phép (tra user_content_rentals.expiry_date); (6) nếu user có vé daily pass và nội dung hợp lệ -> cho phép và trừ vé; (7) nếu tất cả đều không -> từ chối (403) và yêu cầu thanh toán hoặc chờ. Mô hình dữ liệu đã hỗ trợ đầy đủ cho các bước này (cờ is_public trong chapter/episode, bảng purchases, rentals, subscriptions, donation/daily pass nếu có).

***Sơ đồ ER của Hệ thống CSDL***

Dưới đây là sơ đồ ER chi tiết thể hiện các thực thể và quan hệ trong hệ thống, bao gồm các bảng cho Anime, Manga, Novel và bảng dùng chung, kèm lược đồ các thuộc tính chính và quan hệ (cardinality). Mỗi quan hệ khóa ngoại được ký hiệu với “1” nối sang thực thể cha và “N” sang thực thể con (phía nhiều). Các bảng liên kết phục vụ quan hệ nhiều-nhiều được thể hiện rõ (ví dụ anime_character, content_relation…).

Sơ đồ ER mô tả các thực thể (hình chữ nhật) và quan hệ giữa chúng. Màu sắc phân nhóm: Xanh (Anime), Xanh lá (Manga), Cam (Novel), Tím (bảng chung), Nâu (bảng người dùng & giao dịch). Các đường nối kèm nhãn 1/N thể hiện quan hệ một-nhiều. Đường nét đứt biểu thị quan hệ tùy chọn (ví dụ voice_actor trong anime_character có thể null).

***DDL PostgreSQL – Định nghĩa Các Bảng và Ràng buộc***

Dưới đây là DDL SQL chi tiết cho toàn bộ các bảng đã thiết kế, bao gồm khai báo UUID cho khóa chính/ ngoại, các ràng buộc ENUM, INDEX, UNIQUE và FOREIGN KEY, cũng như phần COMMENT mô tả ý nghĩa từng bảng/cột:

```sql
-- Định nghĩa các enum dùng trong hệ thống
CREATE TYPE content_status AS ENUM ('ONGOING','COMPLETED','HIATUS');  -- trạng thái nội dung
CREATE TYPE season_name AS ENUM ('SPRING','SUMMER','FALL','WINTER');  -- mùa phát sóng anime
CREATE TYPE creator_role AS ENUM ('AUTHOR','ILLUSTRATOR','ARTIST','STUDIO','VOICE_ACTOR');  -- vai trò creator
CREATE TYPE content_type_enum AS ENUM ('ANIME','MANGA','NOVEL');  -- loại nội dung
CREATE TYPE content_relation_type AS ENUM ('ADAPTATION','SEQUEL','SPINOFF','RELATED');  -- loại liên kết nội dung
CREATE TYPE purchase_item_type AS ENUM
    ('ANIME_EPISODE','ANIME_SEASON',
     'MANGA_CHAPTER','MANGA_VOLUME','MANGA_SERIES',
     'NOVEL_CHAPTER','NOVEL_VOLUME','NOVEL_SERIES');  -- loại nội dung có thể mua
CREATE TYPE rental_item_type AS ENUM
    ('ANIME_SEASON','ANIME_SERIES',
     'MANGA_VOLUME','MANGA_SERIES',
     'NOVEL_VOLUME','NOVEL_SERIES');  -- loại nội dung cho thuê
CREATE TYPE subscription_tier AS ENUM ('FREE','PREMIUM','VIP');  -- hạng thuê bao

-- Bảng Anime (danh sách series anime)
CREATE TABLE anime (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,         -- trạng thái: đang chiếu, hoàn thành, tạm ngưng
    cover_image TEXT,                       -- URL ảnh bìa
    broadcast_season season_name,           -- mùa phát sóng đầu tiên (Xuân/Hạ/Thu/Đông)
    broadcast_year INT                      -- năm phát sóng đầu tiên
);
COMMENT ON TABLE anime IS 'Anime series master table, one record per anime title.';
COMMENT ON COLUMN anime.status IS 'Current status of the anime (ongoing, completed, etc).';
COMMENT ON COLUMN anime.broadcast_season IS 'Broadcast season (quarter) of the anime''s premiere (Spring, Summer, Fall, Winter).';
COMMENT ON COLUMN anime.broadcast_year IS 'Broadcast year of the anime''s premiere.';

-- Bảng AnimeSeason (các mùa của anime)
CREATE TABLE anime_season (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    season_number INT NOT NULL,       -- số thứ tự mùa trong anime (1,2,...)
    season_title TEXT,               -- tiêu đề mùa (nếu có, ví dụ "Final Season")
    price_coins INT,                 -- giá mua trọn mùa (coins)
    rental_price_coins INT,          -- giá thuê mùa
    rental_duration_days INT,        -- thời gian thuê (ngày)
    UNIQUE(anime_id, season_number)
);
COMMENT ON TABLE anime_season IS 'Seasons of an anime series (e.g. Season 1, Season 2) grouping episodes.';
COMMENT ON COLUMN anime_season.season_number IS 'Sequential season number within the anime series.';
COMMENT ON COLUMN anime_season.price_coins IS 'Price (in virtual coins) to purchase this entire season permanently.';
COMMENT ON COLUMN anime_season.rental_price_coins IS 'Price (in coins) to rent this season for a limited time period.';
COMMENT ON COLUMN anime_season.rental_duration_days IS 'Duration of rental access for this season (in days).';

-- Bảng AnimeEpisode (các tập phim)
CREATE TABLE anime_episode (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    season_id UUID NOT NULL REFERENCES anime_season(id) ON DELETE CASCADE,
    episode_number INT NOT NULL,     -- số tập trong mùa
    title TEXT,                      -- tiêu đề tập (nếu có)
    duration_seconds INT,            -- độ dài tập (giây)
    video_url TEXT,                  -- URL video stream
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí hay không
    price_coins INT,                 -- giá mua lẻ tập
    UNIQUE(season_id, episode_number)
);
COMMENT ON TABLE anime_episode IS 'Individual episodes within an anime season.';
COMMENT ON COLUMN anime_episode.episode_number IS 'Episode number within the season.';
COMMENT ON COLUMN anime_episode.duration_seconds IS 'Duration of the episode in seconds.';
COMMENT ON COLUMN anime_episode.is_public IS 'Whether this episode is freely accessible (public) or requires purchase.';
COMMENT ON COLUMN anime_episode.price_coins IS 'Price (in coins) to purchase this episode.';

-- Bảng EpisodeSubtitle (phụ đề cho tập phim)
CREATE TABLE episode_subtitle (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    episode_id UUID NOT NULL REFERENCES anime_episode(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,  -- mã ngôn ngữ (ví dụ 'en', 'vi')
    subtitle_url TEXT,                   -- URL file phụ đề (.srt, .vtt)
    UNIQUE(episode_id, language_code)
);
COMMENT ON TABLE episode_subtitle IS 'Subtitle file reference for an episode in various languages.';
COMMENT ON COLUMN episode_subtitle.language_code IS 'Language code for the subtitle track (e.g., en, vi, jp).';

-- Bảng Manga (danh sách series manga)
CREATE TABLE manga (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,   -- trạng thái xuất bản
    cover_image TEXT                  -- URL ảnh bìa manga
);
COMMENT ON TABLE manga IS 'Manga series master table (comic).';
COMMENT ON COLUMN manga.status IS 'Current publication status of the manga (ongoing, completed, hiatus).';

-- Bảng MangaVolume (tập truyện của manga)
CREATE TABLE manga_volume (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    volume_number INT NOT NULL,      -- số tập (nếu manga có phân tập)
    volume_title TEXT,               -- tên tập (nếu có)
    cover_image TEXT,                -- ảnh bìa tập
    description TEXT,                -- mô tả ngắn về tập
    price_coins INT,                 -- giá mua trọn tập
    rental_price_coins INT,          -- giá thuê tập
    rental_duration_days INT,        -- thời hạn thuê tập (ngày)
    UNIQUE(manga_id, volume_number)
);
COMMENT ON TABLE manga_volume IS 'Volume grouping chapters of a manga series.';
COMMENT ON COLUMN manga_volume.volume_number IS 'Volume number (if official; a default volume may be used for uncollected chapters).';
COMMENT ON COLUMN manga_volume.price_coins IS 'Price to purchase this volume permanently.';
COMMENT ON COLUMN manga_volume.rental_price_coins IS 'Price to rent this volume temporarily.';
COMMENT ON COLUMN manga_volume.rental_duration_days IS 'Rental duration for this volume (in days).';

-- Bảng MangaChapter (các chương của manga)
CREATE TABLE manga_chapter (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    volume_id UUID NOT NULL REFERENCES manga_volume(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,    -- số thứ tự chương trong tập
    title TEXT,                     -- tiêu đề chương (nếu có)
    released_at TIMESTAMP,          -- thời điểm phát hành chương
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí không
    price_coins INT,               -- giá mua lẻ chương
    UNIQUE(volume_id, chapter_number)
);
COMMENT ON TABLE manga_chapter IS 'Chapter of a manga, containing multiple pages.';
COMMENT ON COLUMN manga_chapter.chapter_number IS 'Chapter number within the volume (or within the series if no formal volumes).';
COMMENT ON COLUMN manga_chapter.released_at IS 'Release timestamp of this chapter.';
COMMENT ON COLUMN manga_chapter.is_public IS 'Whether this chapter is free to read (public) or requires purchase.';
COMMENT ON COLUMN manga_chapter.price_coins IS 'Price to purchase this chapter.';

-- Bảng MangaPage (trang truyện của chapter)
CREATE TABLE manga_page (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    chapter_id UUID NOT NULL REFERENCES manga_chapter(id) ON DELETE CASCADE,
    page_number INT NOT NULL,      -- số thứ tự trang
    image_url TEXT,               -- URL ảnh của trang
    UNIQUE(chapter_id, page_number)
);
COMMENT ON TABLE manga_page IS 'Individual page (image) of a manga chapter.';
COMMENT ON COLUMN manga_page.page_number IS 'Page number within the chapter (for ordering).';
COMMENT ON COLUMN manga_page.image_url IS 'URL or file path of the image for this page.';

-- Bảng Novel (danh sách series novel)
CREATE TABLE novel (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,   -- trạng thái (ongoing, completed, etc.)
    cover_image TEXT                  -- ảnh bìa novel
);
COMMENT ON TABLE novel IS 'Novel (text story) series master table.';
COMMENT ON COLUMN novel.status IS 'Current status of the novel (ongoing, completed, etc).';

-- Bảng NovelVolume (tập của novel)
CREATE TABLE novel_volume (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    volume_number INT NOT NULL,   -- số tập
    volume_title TEXT,            -- tên tập (nếu có)
    cover_image TEXT,             -- ảnh bìa tập
    description TEXT,             -- mô tả ngắn tập
    price_coins INT,              -- giá mua trọn tập
    rental_price_coins INT,       -- giá thuê tập
    rental_duration_days INT,     -- thời hạn thuê tập (ngày)
    UNIQUE(novel_id, volume_number)
);
COMMENT ON TABLE novel_volume IS 'Volume grouping chapters of a novel.';
COMMENT ON COLUMN novel_volume.volume_number IS 'Volume number in the novel series.';
COMMENT ON COLUMN novel_volume.price_coins IS 'Price to purchase this volume.';
COMMENT ON COLUMN novel_volume.rental_price_coins IS 'Price to rent this volume.';
COMMENT ON COLUMN novel_volume.rental_duration_days IS 'Rental duration for this volume (days).';

-- Bảng NovelChapter (các chương của novel)
CREATE TABLE novel_chapter (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    volume_id UUID NOT NULL REFERENCES novel_volume(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,   -- số chương trong tập
    title TEXT,                    -- tiêu đề chương
    content TEXT,                  -- nội dung văn bản của chương
    published_at TIMESTAMP,        -- thời gian xuất bản chương
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí không
    price_coins INT,               -- giá mua lẻ chương
    UNIQUE(volume_id, chapter_number)
);
COMMENT ON TABLE novel_chapter IS 'Chapter of a novel, containing text content.';
COMMENT ON COLUMN novel_chapter.chapter_number IS 'Chapter number within the volume.';
COMMENT ON COLUMN novel_chapter.content IS 'Full text content of the chapter.';
COMMENT ON COLUMN novel_chapter.published_at IS 'Publication timestamp of the chapter.';
COMMENT ON COLUMN novel_chapter.is_public IS 'Whether this chapter is free (public) or requires purchase.';
COMMENT ON COLUMN novel_chapter.price_coins IS 'Price to purchase this chapter.';

-- Bảng Character (nhân vật)
CREATE TABLE character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,        -- tên nhân vật
    description TEXT,          -- mô tả
    image_url TEXT             -- ảnh đại diện nhân vật
);
COMMENT ON TABLE character IS 'Fictional character that can appear in multiple content (anime/manga/novel).';
COMMENT ON COLUMN character.name IS 'Name of the character.';

-- Bảng Creator (nhà sáng tạo: tác giả, họa sĩ, studio, seiyuu, v.v.)
CREATE TABLE creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,       -- tên creator (tên người hoặc tên nhóm/studio)
    description TEXT
);
COMMENT ON TABLE creator IS 'Content creator (author, artist, studio, voice actor, etc.) entity.';
COMMENT ON COLUMN creator.name IS 'Name of the creator (person or organization).';

-- Bảng Genre (thể loại nội dung)
CREATE TABLE genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL        -- tên thể loại
);
COMMENT ON TABLE genre IS 'Genre/category for content (shared across anime, manga, novel).';
COMMENT ON COLUMN genre.name IS 'Name of the genre.';

-- Bảng anime_character (liên kết anime và nhân vật, kèm diễn viên lồng tiếng)
CREATE TABLE anime_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    voice_actor_id UUID REFERENCES creator(id) ON DELETE SET NULL,  -- diễn viên lồng tiếng (có thể null nếu chưa biết)
    UNIQUE(anime_id, character_id)
);
COMMENT ON TABLE anime_character IS 'Maps characters to anime appearances (with optional voice actor casting).';
COMMENT ON COLUMN anime_character.voice_actor_id IS 'Voice actor (creator) who voices the character in this anime.';

-- Bảng manga_character (liên kết manga và nhân vật)
CREATE TABLE manga_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    UNIQUE(manga_id, character_id)
);
COMMENT ON TABLE manga_character IS 'Maps characters to manga appearances.';

-- Bảng novel_character (liên kết novel và nhân vật)
CREATE TABLE novel_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    UNIQUE(novel_id, character_id)
);
COMMENT ON TABLE novel_character IS 'Maps characters to novel appearances.';

-- Bảng anime_genre (liên kết anime và thể loại)
CREATE TABLE anime_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(anime_id, genre_id)
);
COMMENT ON TABLE anime_genre IS 'Associates anime with genres (many-to-many).';

-- Bảng manga_genre (liên kết manga và thể loại)
CREATE TABLE manga_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(manga_id, genre_id)
);
COMMENT ON TABLE manga_genre IS 'Associates manga with genres.';

-- Bảng novel_genre (liên kết novel và thể loại)
CREATE TABLE novel_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(novel_id, genre_id)
);
COMMENT ON TABLE novel_genre IS 'Associates novels with genres.';

-- Bảng anime_creator (liên kết anime và creator với vai trò)
CREATE TABLE anime_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(anime_id, creator_id, role)
);
COMMENT ON TABLE anime_creator IS 'Associates anime with creators (e.g., studio production company).';

-- Bảng manga_creator (liên kết manga và creator)
CREATE TABLE manga_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(manga_id, creator_id, role)
);
COMMENT ON TABLE manga_creator IS 'Associates manga with creators (e.g., author, artist).';

-- Bảng novel_creator (liên kết novel và creator)
CREATE TABLE novel_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(novel_id, creator_id, role)
);
COMMENT ON TABLE novel_creator IS 'Associates novel with creators (e.g., author, illustrator).';

-- Bảng content_relation (liên kết các nội dung với nhau: adaptation, sequel, etc.)
CREATE TABLE content_relation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    source_id UUID NOT NULL,
    source_type content_type_enum NOT NULL,
    target_id UUID NOT NULL,
    target_type content_type_enum NOT NULL,
    relation_type content_relation_type NOT NULL
    -- Không đặt FOREIGN KEY trực tiếp do source/target có thể là nhiều bảng khác nhau
);
COMMENT ON TABLE content_relation IS 'Inter-content relationships between works (adaptation, sequel, spinoff, etc).';
COMMENT ON COLUMN content_relation.source_type IS 'Type of source content (ANIME, MANGA, NOVEL).';
COMMENT ON COLUMN content_relation.relation_type IS 'Relationship type (adaptation, sequel, spinoff, related).';

-- Index hỗ trợ tra cứu content_relation
CREATE INDEX idx_content_relation_source ON content_relation(source_type, source_id);
CREATE INDEX idx_content_relation_target ON content_relation(target_type, target_id);

-- Bảng anime_translation (tiêu đề/mô tả đa ngôn ngữ cho anime)
CREATE TABLE anime_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,  -- mã ngôn ngữ (ví dụ 'en', 'vi')
    title TEXT NOT NULL,                -- tiêu đề anime bằng ngôn ngữ này
    description TEXT,                   -- mô tả bằng ngôn ngữ này
    is_primary BOOLEAN DEFAULT FALSE,   -- đánh dấu tên chính
    UNIQUE(anime_id, language_code, title)  -- tránh trùng exact tiêu đề (có thể bỏ nếu cho phép)
);
-- Đảm bảo mỗi anime mỗi ngôn ngữ chỉ có một title chính thức
CREATE UNIQUE INDEX ux_anime_translation_primary ON anime_translation(anime_id, language_code) WHERE is_primary;
COMMENT ON TABLE anime_translation IS 'Localized titles and descriptions for anime in multiple languages.';
COMMENT ON COLUMN anime_translation.language_code IS 'Locale/language code of this translation.';
COMMENT ON COLUMN anime_translation.is_primary IS 'Indicates the primary title for the anime in this language.';

-- Bảng manga_translation (đa ngôn ngữ cho manga)
CREATE TABLE manga_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    UNIQUE(manga_id, language_code, title)
);
CREATE UNIQUE INDEX ux_manga_translation_primary ON manga_translation(manga_id, language_code) WHERE is_primary;
COMMENT ON TABLE manga_translation IS 'Localized titles and descriptions for manga in multiple languages.';

-- Bảng novel_translation (đa ngôn ngữ cho novel)
CREATE TABLE novel_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    UNIQUE(novel_id, language_code, title)
);
CREATE UNIQUE INDEX ux_novel_translation_primary ON novel_translation(novel_id, language_code) WHERE is_primary;
COMMENT ON TABLE novel_translation IS 'Localized titles and descriptions for novels in multiple languages.';

-- Bảng user_content_purchases (lịch sử mua nội dung của user)
CREATE TABLE user_content_purchases (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,         -- tham chiếu user (từ service user)
    item_type purchase_item_type NOT NULL,  -- loại nội dung đã mua
    item_id UUID NOT NULL,         -- id của nội dung (tập/chương/volume...)
    purchase_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- (Không dùng FK cứng cho item_id vì có nhiều loại nội dung, xác thực qua ứng dụng)
);
COMMENT ON TABLE user_content_purchases IS 'Records of user purchases of content (episodes, chapters, volumes, etc).';
COMMENT ON COLUMN user_content_purchases.user_id IS 'ID of the user who made the purchase (from user service).';
COMMENT ON COLUMN user_content_purchases.item_type IS 'Type of content item purchased (e.g. NOVEL_CHAPTER, MANGA_VOLUME).';
COMMENT ON COLUMN user_content_purchases.item_id IS 'ID of the purchased content item.';
CREATE INDEX idx_user_purchase_user ON user_content_purchases(user_id);
CREATE INDEX idx_user_purchase_item ON user_content_purchases(item_type, item_id, user_id);

-- Bảng user_content_rentals (lịch sử thuê nội dung của user)
CREATE TABLE user_content_rentals (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    item_type rental_item_type NOT NULL,  -- loại nội dung thuê
    item_id UUID NOT NULL,
    rent_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date TIMESTAMP NOT NULL       -- thời điểm hết hạn thuê
);
COMMENT ON TABLE user_content_rentals IS 'Records of user rentals of content with limited-time access.';
COMMENT ON COLUMN user_content_rentals.expiry_date IS 'Datetime when the rental access expires for the user.';
CREATE INDEX idx_user_rental_user ON user_content_rentals(user_id);
CREATE INDEX idx_user_rental_item ON user_content_rentals(item_type, item_id, user_id);

-- Bảng user_subscriptions (thuê bao của user)
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    tier subscription_tier NOT NULL,   -- hạng thuê bao (FREE/PREMIUM/VIP)
    start_date DATE NOT NULL,
    end_date DATE                     -- ngày hết hạn (null nếu đang active không kỳ hạn cố định)
);
COMMENT ON TABLE user_subscriptions IS 'User subscriptions (membership plans) for premium access/perks.';
COMMENT ON COLUMN user_subscriptions.tier IS 'Subscription plan tier (e.g., VIP for special perks).';
COMMENT ON COLUMN user_subscriptions.end_date IS 'End date of the subscription (if applicable).';
CREATE INDEX idx_user_subscription_user ON user_subscriptions(user_id);

-- Bảng user_donations (lịch sử donate của user)
CREATE TABLE user_donations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    content_type content_type_enum NOT NULL,  -- loại nội dung được donate (ANIME/MANGA/NOVEL)
    content_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,           -- số tiền donate (giả sử đơn vị USD hoặc coin)
    donation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE user_donations IS 'Records of users donating to content creators (by content).';
COMMENT ON COLUMN user_donations.amount IS 'Donation amount (could be in platform currency).';
CREATE INDEX idx_user_donation_user ON user_donations(user_id);
CREATE INDEX idx_user_donation_content ON user_donations(content_type, content_id);
```

***Một số điểm cần lưu ý trong DDL trên:***
	•	Tất cả khóa chính (id) đều dùng kiểu UUID với default uuidv7() để PostgreSQL sinh giá trị có tính đơn điệu theo thời gian (phù hợp với Identify Service). Hãy đảm bảo hàm uuidv7() được triển khai sẵn (vd. thông qua extension hoặc migration chia sẻ) trước khi chạy DDL; cách này vẫn giữ được tính duy nhất toàn cục và giảm rò rỉ ID so với sequence.
	•	Các FOREIGN KEY được thiết lập với chiến lược xóa hợp lý: hầu hết dùng ON DELETE CASCADE khi quan hệ cha-con (xóa cha xóa con, ví dụ xóa anime xóa luôn season, episode liên quan). Các trường tham chiếu đối tượng dùng chung (nhân vật, creator) thì thường dùng CASCADE nếu xóa đối tượng đó (xóa nhân vật sẽ xóa luôn liên kết trong anime_character, vì nếu nhân vật không tồn tại thì quan hệ cũng vô nghĩa). Trường hợp đặc biệt: voice_actor_id trong anime_character dùng ON DELETE SET NULL – nếu một creator diễn viên bị xóa, ta không muốn mất bản ghi anime_character (vì nhân vật vẫn xuất hiện trong anime) nên chỉ đặt voice_actor_id = null (chưa có diễn viên).
	•	Nhiều UNIQUE constraints được thêm để đảm bảo tính toàn vẹn logic: ví dụ (anime_id, season_number), (volume_id, chapter_number),… giúp ngăn dữ liệu trùng lặp bất hợp lệ. Những chỗ cần cặp nhiều field làm khóa chính logic nhưng ta vẫn dùng UUID làm khóa chính vật lý thì unique index phục vụ kiểm tra là đủ.
	•	Các INDEX phi khóa chính: Đã thêm index cho các cột thường dùng tra cứu: user_id trong bảng purchases/rentals/subscriptions/donations để nhanh chóng lấy danh sách giao dịch của một user; index ghép (item_type, item_id) trong purchases/rentals để tra cứu nhanh xem một user đã mua/thuê nội dung cụ thể chưa (kết hợp với user_id trong mệnh đề WHERE); index cho content_relation theo source và target để truy vấn nội dung liên quan. Ngoài ra, index một phần cho bảng translation (ux_translation_primary) để đảm bảo nhanh trong kiểm tra tên chính.
	•	Các COMMENT được thêm cho mọi bảng và nhiều cột để giải thích ý nghĩa. Điều này giúp tự mô tả (self-documenting) schema, rất hữu ích khi làm việc nhóm hoặc mở API cho bên thứ ba.

Như vậy, hệ thống cơ sở dữ liệu được thiết kế ở mức chi tiết, đáp ứng đầy đủ các yêu cầu: tách biệt rõ ba phân hệ Anime/Manga/Novel với cấu trúc chuyên biệt nhưng có thành phần dùng chung (Character, Creator, Genre, Link) để kết nối liền mạch nội dung đa dạng; hỗ trợ đa ngôn ngữ cho metadata nội dung; và tích hợp chặt chẽ mô hình kinh doanh linh hoạt (mua lẻ, thuê, thuê bao, daily pass, donate) với các bảng giao dịch và cờ kiểm soát tương ứng.

Hệ thống sử dụng UUID làm khóa chính/xuất ngoại cho mọi bảng (kể cả user_id từ service ngoài) nhằm đảm bảo tính duy nhất và an toàn. Thiết kế này hướng đến khả năng mở rộng (thêm nội dung mới, thêm ngôn ngữ, thêm mô hình kinh doanh) một cách dễ dàng, đồng thời đảm bảo rõ ràng trong cấu trúc và toàn vẹn dữ liệu thông qua các ràng buộc. Với mô hình này, nền tảng có thể phục vụ một cộng đồng lớn, nơi người dùng có thể xem anime, đọc truyện tranh, tiểu thuyết liền mạch, khám phá nhân vật chung, và tận hưởng các tiện ích như bảng xếp hạng, bình luận, đánh giá… trên một cơ sở dữ liệu thống nhất, vững chắc.
