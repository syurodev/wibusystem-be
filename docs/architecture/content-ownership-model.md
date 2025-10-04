# Mô hình Quản lý Quyền Sở hữu Nội dung - WibuSystem

**Phiên bản:** 1.0
**Cập nhật:** 04/10/2025
**Trạng thái:** Tài liệu Thiết kế Nghiệp vụ

---

## 📋 Mục lục

1. [Tổng quan](#1-tổng-quan)
2. [Bối cảnh và Vấn đề](#2-bối-cảnh-và-vấn-đề)
3. [Yêu cầu Nghiệp vụ](#3-yêu-cầu-nghiệp-vụ)
4. [Các Loại Quyền Sở hữu](#4-các-loại-quyền-sở-hữu)
5. [Kịch bản Người dùng](#5-kịch-bản-người-dùng)
6. [Quy trình Chuyển giao Nội dung](#6-quy-trình-chuyển-giao-nội-dung)
7. [Phân quyền và Truy cập](#7-phân-quyền-và-truy-cập)
8. [Câu hỏi Thường gặp](#8-câu-hỏi-thường-gặp)
9. [Lộ trình Triển khai](#9-lộ-trình-triển-khai)
10. [Phụ lục](#10-phụ-lục)

---

## 1. Tổng quan

### 1.1 WibuSystem là gì?

WibuSystem là nền tảng SaaS (Software as a Service) cho phép:

- **Tác giả cá nhân** đăng tải và quản lý truyện của riêng mình
- **Nhà xuất bản / Tổ chức (Tenant)** quản lý nội dung theo nhóm
- **Linh hoạt chuyển đổi** quyền sở hữu khi tác giả gia nhập tổ chức

### 1.2 Mục đích Tài liệu

Tài liệu này mô tả cách thức quản lý quyền sở hữu nội dung trong hệ thống, bao gồm:

- Ai là chủ sở hữu của một tác phẩm?
- Điều gì xảy ra khi tác giả gia nhập một nhà xuất bản?
- Làm thế nào để chuyển giao quyền sở hữu?
- Ai có quyền làm gì với nội dung?

### 1.3 Đối tượng Đọc

- **Quản lý Sản phẩm**: Hiểu rõ tính năng để định hướng phát triển
- **Đội ngũ Kinh doanh**: Tư vấn khách hàng về mô hình hoạt động
- **Đội ngũ Hỗ trợ**: Giải quyết thắc mắc của người dùng
- **Đội ngũ Phát triển**: Tham khảo để hiểu nghiệp vụ trước khi implement

---

## 2. Bối cảnh và Vấn đề

### 2.1 Bối cảnh Thị trường

Thị trường xuất bản truyện trực tuyến có 3 nhóm người dùng chính:

**1. Tác giả nghiệp dư (Hobbyist Writers)**

- Viết truyện như sở thích cá nhân
- Muốn kiểm soát hoàn toàn tác phẩm của mình
- Không muốn ràng buộc với tổ chức nào

**2. Tác giả chuyên nghiệp (Professional Authors)**

- Bắt đầu viết cá nhân, sau đó được nhà xuất bản mời hợp tác
- Muốn chuyển một số tác phẩm cho nhà xuất bản quản lý
- Vẫn giữ quyền sở hữu một số tác phẩm khác

**3. Nhà xuất bản / Công ty truyện (Publishers)**

- Quản lý nhiều tác giả và tác phẩm
- Cần quyền kiểm soát nội dung do công ty sở hữu
- Muốn hợp tác với tác giả độc lập

### 2.2 Vấn đề Cần Giải quyết

**Câu hỏi 1:** Một tác giả đã đăng 50 truyện cá nhân. Giờ họ gia nhập nhà xuất bản A. Điều gì xảy ra với 50 truyện đó?

**Câu hỏi 2:** Tác giả muốn chuyển 20 truyện cho nhà xuất bản A, nhưng giữ lại 30 truyện cá nhân. Có được không?

**Câu hỏi 3:** Tác giả vừa làm việc cho nhà xuất bản A, vừa viết truyện riêng, vừa cộng tác với nhà xuất bản B. Hệ thống quản lý thế nào?

**Câu hỏi 4:** Tác giả rời khỏi nhà xuất bản. Những truyện đã chuyển giao có quay lại tác giả không?

**Câu hỏi 5:** Làm sao theo dõi được ai là người sáng tạo ban đầu, dù tác phẩm đã đổi chủ nhiều lần?

### 2.3 Giải pháp Đề xuất

Hệ thống sẽ hỗ trợ **3 loại quyền sở hữu linh hoạt**:

1. **Sở hữu Cá nhân** - Tác giả làm chủ hoàn toàn
2. **Sở hữu Tổ chức** - Nhà xuất bản làm chủ hoàn toàn
3. **Sở hữu Hợp tác** - Tác giả và Nhà xuất bản cùng quản lý

---

## 3. Yêu cầu Nghiệp vụ

### 3.1 Yêu cầu Bắt buộc (Must Have)

| Mã        | Yêu cầu                                              | Lý do                                               |
| --------- | ---------------------------------------------------- | --------------------------------------------------- |
| **YC-01** | Tác giả có thể tạo và sở hữu truyện cá nhân          | Tác giả độc lập cần không gian riêng để sáng tạo    |
| **YC-02** | Tác giả có thể gia nhập nhiều Tổ chức (Tenant)       | Một tác giả có thể làm việc cho nhiều nhà xuất bản  |
| **YC-03** | Tác giả có thể chuyển truyện từ cá nhân sang Tổ chức | Khi ký hợp đồng, tác giả muốn chuyển giao quyền     |
| **YC-04** | Hệ thống lưu lại lịch sử chuyển giao                 | Để giải quyết tranh chấp và kiểm tra                |
| **YC-05** | Luôn biết được ai là tác giả gốc                     | Bảo vệ quyền sáng tạo dù đổi chủ nhiều lần          |
| **YC-06** | Phân quyền rõ ràng cho từng loại sở hữu              | Tránh xung đột quyền hạn                            |
| **YC-07** | Xử lý trường hợp tác giả rời Tổ chức                 | Quyết định số phận của truyện khi tác giả nghỉ việc |

### 3.2 Yêu cầu Nên có (Should Have)

| Mã        | Yêu cầu                                           | Lý do                                           |
| --------- | ------------------------------------------------- | ----------------------------------------------- |
| **YC-08** | Chuyển giao qua quy trình phê duyệt               | Nhà xuất bản muốn xem xét trước khi nhận truyện |
| **YC-09** | Tác giả có thể hợp tác với Tổ chức (co-ownership) | Mô hình chia sẻ doanh thu hiện đại              |
| **YC-10** | Tổ chức có thể chuyển truyện lại cho tác giả      | Khi hợp đồng kết thúc, trả quyền lại tác giả    |
| **YC-11** | Báo cáo thống kê theo loại sở hữu                 | Giúp quản lý theo dõi tài sản nội dung          |

### 3.3 Yêu cầu Có thể có (Nice to Have)

| Mã        | Yêu cầu                                      | Lý do                                            |
| --------- | -------------------------------------------- | ------------------------------------------------ |
| **YC-12** | Chuyển giao hàng loạt nhiều truyện cùng lúc  | Tiết kiệm thời gian khi chuyển giao số lượng lớn |
| **YC-13** | Thông báo tự động khi có yêu cầu chuyển giao | Tăng tốc độ phê duyệt                            |
| **YC-14** | Mẫu hợp đồng chuyển giao tự động             | Hỗ trợ pháp lý cho cả hai bên                    |

---

## 4. Các Loại Quyền Sở hữu

### 4.1 So sánh 3 Loại Sở hữu

| Tiêu chí                 | **Cá nhân**               | **Tổ chức**                   | **Hợp tác**                    |
| ------------------------ | ------------------------- | ----------------------------- | ------------------------------ |
| **Chủ sở hữu chính**     | Tác giả                   | Nhà xuất bản                  | Tác giả (nhưng chia sẻ quyền)  |
| **Thuộc Tổ chức?**       | Không                     | Có                            | Có                             |
| **Ai quản lý?**          | Chỉ tác giả               | Nhà xuất bản + biên tập viên  | Cả tác giả và nhà xuất bản     |
| **Ai xem được?**         | Chỉ tác giả               | Mọi thành viên Tổ chức        | Cả hai bên                     |
| **Ai sửa được?**         | Chỉ tác giả               | Nhà xuất bản + biên tập viên  | Cả hai bên                     |
| **Ai xóa được?**         | Chỉ tác giả               | Quản trị viên Tổ chức         | Cả hai bên (phải đồng ý)       |
| **Ai chuyển giao được?** | Chỉ tác giả               | Quản trị viên Tổ chức         | Cả hai bên                     |
| **Ví dụ thực tế**        | Truyện fanfiction cá nhân | Truyện độc quyền nhà xuất bản | Truyện chia sẻ doanh thu 50-50 |

### 4.2 Loại 1: Sở hữu Cá nhân (Personal)

**Định nghĩa:**

> Truyện do tác giả cá nhân tạo ra và quản lý hoàn toàn, không thuộc về bất kỳ tổ chức nào.

**Đặc điểm:**

- ✅ Tác giả có quyền tuyệt đối: xem, sửa, xóa, chuyển giao
- ✅ Không ai khác xem được (trừ khi tác giả công khai)
- ✅ Tác giả có thể chuyển sang Tổ chức hoặc Hợp tác bất kỳ lúc nào
- ❌ Thành viên Tổ chức KHÔNG xem được

**Ví dụ:**

- Nguyễn Văn A viết truyện "Rừng Na Uy" như sở thích cá nhân
- Anh ấy chưa muốn gia nhập nhà xuất bản nào
- Chỉ anh ấy mới thấy và chỉnh sửa truyện này

### 4.3 Loại 2: Sở hữu Tổ chức (Tenant)

**Định nghĩa:**

> Truyện thuộc sở hữu của nhà xuất bản/tổ chức, được quản lý bởi đội ngũ của tổ chức.

**Đặc điểm:**

- ✅ Mọi thành viên Tổ chức (theo phân quyền) đều xem được
- ✅ Biên tập viên có thể sửa nội dung
- ✅ Quản trị viên có thể chuyển giao hoặc xóa
- ❌ Tác giả gốc KHÔNG tự động có quyền (trừ khi là thành viên Tổ chức)
- ℹ️ Tác giả gốc vẫn được ghi nhận, nhưng không có quyền kiểm soát

**Ví dụ:**

- Trần Thị B chuyển truyện "Mắt Biếc" cho Nhà xuất bản Kim Đồng
- Kim Đồng trở thành chủ sở hữu, quản lý toàn bộ
- Nếu Trần Thị B rời công ty, truyện vẫn ở lại với Kim Đồng
- Nhưng tên Trần Thị B vẫn được ghi là "Tác giả gốc"

### 4.4 Loại 3: Sở hữu Hợp tác (Collaborative)

**Định nghĩa:**

> Truyện được chia sẻ quyền sở hữu giữa tác giả cá nhân và tổ chức, cả hai cùng quản lý.

**Đặc điểm:**

- ✅ Tác giả vẫn là chủ sở hữu chính
- ✅ Tổ chức có quyền quản lý và phát hành
- ✅ Cả hai bên đều có thể xem và sửa
- ⚠️ Quyết định quan trọng (xóa, chuyển giao) cần cả hai bên đồng ý
- ℹ️ Phù hợp với mô hình chia sẻ doanh thu

**Ví dụ:**

- Lê Văn C ký hợp đồng chia sẻ với Nhà xuất bản Trẻ
- Truyện "Tôi thấy hoa vàng" vẫn thuộc quyền sở hữu của Lê Văn C
- Nhưng Nhà xuất bản Trẻ được phép phát hành và marketing
- Doanh thu chia 60% (tác giả) - 40% (nhà xuất bản)
- Nếu Lê Văn C muốn xóa truyện, cần thỏa thuận với Nhà xuất bản Trẻ

---

## 5. Kịch bản Người dùng

### 5.1 Kịch bản 1: Tác giả Độc lập

**Nhân vật:** Mai - Sinh viên năm 3, thích viết truyện

**Tình huống:**

- Mai viết truyện fanfiction về anime yêu thích
- Cô ấy chỉ muốn chia sẻ với bạn bè, không muốn thương mại hóa

**Hành động:**

1. Mai đăng ký tài khoản WibuSystem
2. Mai tạo truyện mới "Nhật ký của Sakura"
3. Hệ thống tự động đặt: **Sở hữu Cá nhân**
4. Truyện chỉ Mai nhìn thấy (trừ khi Mai công khai)

**Kết quả:**

- ✅ Mai hoàn toàn kiểm soát truyện của mình
- ✅ Không ai can thiệp vào nội dung
- ✅ Mai có thể xóa hoặc chuyển giao bất cứ lúc nào

---

### 5.2 Kịch bản 2: Tác giả Gia nhập Nhà xuất bản

**Nhân vật:** Hùng - Tác giả có 20 truyện cá nhân

**Tình huống:**

- Hùng được Nhà xuất bản "Văn Học Việt" mời hợp tác
- Hùng muốn chuyển 10 truyện cho nhà xuất bản, giữ lại 10 truyện cá nhân

**Hành động:**

**Bước 1: Gia nhập Tổ chức**

- Hùng nhận lời mời từ "Văn Học Việt"
- Hùng trở thành thành viên với vai trò "Tác giả"

**Bước 2: Xem lại truyện cá nhân**

- Hùng vào mục "Truyện của tôi"
- Hệ thống hiển thị 20 truyện **Sở hữu Cá nhân**

**Bước 3: Chuyển giao truyện**

- Hùng chọn 10 truyện muốn chuyển
- Hùng chọn: "Chuyển cho Tổ chức: Văn Học Việt"
- Hùng chọn loại chuyển giao:
  - **Chuyển giao hoàn toàn** (Văn Học Việt làm chủ) → Chọn 5 truyện
  - **Hợp tác chia sẻ** (Cùng quản lý) → Chọn 5 truyện
- Hùng điền lý do: "Theo hợp đồng số 2025/VHV"
- Hệ thống tạo yêu cầu chuyển giao, chờ phê duyệt

**Bước 4: Quản trị viên phê duyệt**

- Giám đốc "Văn Học Việt" nhận thông báo
- Giám đốc xem chi tiết 10 truyện
- Giám đốc phê duyệt

**Bước 5: Hoàn tất**

- 5 truyện chuyển giao hoàn toàn → **Sở hữu Tổ chức**
  - Hùng KHÔNG còn quyền chỉnh sửa (trừ khi được cấp quyền biên tập viên)
  - Văn Học Việt quản lý hoàn toàn
- 5 truyện hợp tác → **Sở hữu Hợp tác**
  - Hùng vẫn chỉnh sửa được
  - Văn Học Việt cũng chỉnh sửa được
- 10 truyện còn lại → Vẫn là **Sở hữu Cá nhân**

**Kết quả:**

- ✅ Hùng linh hoạt quản lý 3 loại truyện
- ✅ Hệ thống lưu đầy đủ lịch sử chuyển giao
- ✅ Mọi thay đổi đều có thể tra cứu

---

### 5.3 Kịch bản 3: Tác giả Rời Tổ chức

**Nhân vật:** Lan - Tác giả từng ở "Nhà xuất bản Trẻ"

**Tình huống:**

- Lan quyết định nghỉ việc ở "Nhà xuất bản Trẻ"
- Lan có:
  - 3 truyện **Sở hữu Tổ chức** (đã chuyển giao hoàn toàn)
  - 2 truyện **Sở hữu Hợp tác**
  - 5 truyện **Sở hữu Cá nhân**

**Hành động:**

**Bước 1: Lan yêu cầu rời Tổ chức**

- Lan vào "Cài đặt" → "Rời khỏi Nhà xuất bản Trẻ"

**Bước 2: Hệ thống phân tích ảnh hưởng**

- Hệ thống hiển thị:
  - 3 truyện Sở hữu Tổ chức → **Sẽ ở lại** với Nhà xuất bản Trẻ
  - 2 truyện Hợp tác → **Cần quyết định**
  - 5 truyện Cá nhân → **Không ảnh hưởng**

**Bước 3: Lan quyết định với 2 truyện Hợp tác**

| Truyện                | Quyết định của Lan                                    |
| --------------------- | ----------------------------------------------------- |
| "Chuyện Tình Sài Gòn" | Chuyển về Cá nhân (lấy lại quyền sở hữu)              |
| "Mùa Hè Không Độ"     | Để lại cho Nhà xuất bản (chuyển thành Sở hữu Tổ chức) |

**Bước 4: Nhà xuất bản Trẻ phê duyệt**

- Giám đốc xem yêu cầu
- Đồng ý với "Chuyện Tình Sài Gòn" chuyển về Lan
- Đồng ý giữ "Mùa Hè Không Độ"

**Bước 5: Hoàn tất rời Tổ chức**

**Kết quả sau khi Lan rời đi:**

| Truyện                            | Loại sở hữu ban đầu | Loại sở hữu sau khi rời | Ai quản lý?      |
| --------------------------------- | ------------------- | ----------------------- | ---------------- |
| 3 truyện đã chuyển giao hoàn toàn | Tổ chức             | Tổ chức                 | Nhà xuất bản Trẻ |
| "Chuyện Tình Sài Gòn"             | Hợp tác             | Cá nhân                 | Lan              |
| "Mùa Hè Không Độ"                 | Hợp tác             | Tổ chức                 | Nhà xuất bản Trẻ |
| 5 truyện cá nhân                  | Cá nhân             | Cá nhân                 | Lan              |

**Lưu ý quan trọng:**

- ✅ Lan vẫn được ghi nhận là "Tác giả gốc" của tất cả các truyện
- ✅ Lịch sử chuyển giao được lưu lại
- ✅ Nếu có tranh chấp, có thể tra cứu

---

### 5.4 Kịch bản 4: Tác giả Đa Tổ chức

**Nhân vật:** Tuấn - Tác giả làm việc cho 2 nhà xuất bản

**Tình huống:**

- Tuấn là thành viên của:
  - "Nhà xuất bản Kim Đồng" (Vai trò: Tác giả)
  - "Nhà xuất bản Trẻ" (Vai trò: Biên tập viên)
- Tuấn có:
  - 5 truyện cá nhân
  - 3 truyện thuộc Kim Đồng
  - 2 truyện hợp tác với Trẻ

**Hành động:**

**Bước 1: Chọn ngữ cảnh làm việc**

- Tuấn vào WibuSystem
- Tuấn thấy thanh chọn Tổ chức: "Cá nhân | Kim Đồng | Trẻ"
- Tuấn chọn "Kim Đồng"

**Bước 2: Tạo truyện mới**

- Tuấn tạo truyện "Doraemon Việt Nam"
- Hệ thống hỏi: "Tạo cho ai?"
  - Cá nhân của tôi
  - **Nhà xuất bản Kim Đồng** ← Tuấn chọn
  - Nhà xuất bản Trẻ
- Truyện được tạo với **Sở hữu Tổ chức** (thuộc Kim Đồng)

**Bước 3: Xem tổng quan truyện**

- Tuấn vào "Truyện của tôi"
- Hệ thống hiển thị theo nhóm:

```
📚 Truyện Cá nhân (5)
   - Truyện A, B, C, D, E

🏢 Kim Đồng (4)
   - Truyện F, G, H
   - Doraemon Việt Nam (mới tạo)

🏢 Nhà xuất bản Trẻ (2 - Hợp tác)
   - Truyện I (Hợp tác)
   - Truyện J (Hợp tác)
```

**Kết quả:**

- ✅ Tuấn quản lý được nhiều loại truyện khác nhau
- ✅ Rõ ràng từng truyện thuộc về ai
- ✅ Linh hoạt chuyển đổi giữa các Tổ chức

---

## 6. Quy trình Chuyển giao Nội dung

### 6.1 Quy trình Tổng quan

```
[Tác giả có truyện Cá nhân]
          │
          ├─ Tác giả gia nhập Tổ chức
          │
          ▼
[Chọn truyện muốn chuyển giao]
          │
          ├─ Chọn loại chuyển giao:
          │  • Chuyển giao hoàn toàn
          │  • Hợp tác chia sẻ
          │
          ▼
[Yêu cầu chuyển giao được tạo]
          │
          ├─ Trạng thái: Chờ phê duyệt
          │
          ▼
[Quản trị viên Tổ chức xem xét]
          │
          ├─ Phê duyệt ✓
          ├─ Từ chối ✗
          │
          ▼
[Chuyển giao hoàn tất]
          │
          ├─ Quyền sở hữu thay đổi
          ├─ Lịch sử được lưu
          │
          ▼
[Tác giả và Tổ chức nhận thông báo]
```

### 6.2 Chi tiết Từng Bước

#### Bước 1: Khởi tạo Chuyển giao

**Ai thực hiện:** Tác giả (chủ sở hữu hiện tại)

**Thông tin cần có:**

- Chọn truyện (có thể chọn nhiều)
- Chọn Tổ chức đích (phải là Tổ chức mà tác giả đang là thành viên)
- Chọn loại chuyển giao:
  - **Chuyển giao hoàn toàn:** Tổ chức sẽ làm chủ, tác giả mất quyền kiểm soát
  - **Hợp tác chia sẻ:** Cả hai cùng quản lý
- Nhập lý do (tùy chọn): VD "Theo hợp đồng số XYZ"

**Kết quả:**

- Hệ thống tạo "Yêu cầu chuyển giao" với trạng thái **Chờ phê duyệt**
- Gửi thông báo cho quản trị viên Tổ chức

#### Bước 2: Phê duyệt

**Ai thực hiện:** Quản trị viên Tổ chức

**Thông tin xem xét:**

- Danh sách truyện muốn chuyển
- Thông tin tác giả
- Lý do chuyển giao
- Thống kê truyện (lượt xem, đánh giá...)

**Hành động:**

- **Phê duyệt:** Chấp nhận nhận truyện
- **Từ chối:** Không nhận, ghi rõ lý do

**Kết quả:**

- Nếu phê duyệt → Chuyển sang Bước 3
- Nếu từ chối → Tác giả nhận thông báo, truyện vẫn ở dạng cũ

#### Bước 3: Thực hiện Chuyển giao

**Ai thực hiện:** Hệ thống tự động

**Hành động của hệ thống:**

**Nếu là Chuyển giao hoàn toàn:**

- Đổi quyền sở hữu: Cá nhân → Tổ chức
- Tác giả gốc vẫn được ghi nhận
- Lưu lịch sử: "Chuyển từ [Tác giả X] sang [Tổ chức Y] vào [Ngày Z]"

**Nếu là Hợp tác chia sẻ:**

- Đổi quyền sở hữu: Cá nhân → Hợp tác
- Thêm Tổ chức vào danh sách người cộng tác
- Tác giả vẫn giữ quyền chủ sở hữu chính

**Thông báo:**

- Gửi email cho tác giả: "Chuyển giao thành công"
- Gửi thông báo cho Tổ chức: "Đã nhận truyện mới"

### 6.3 Các Trường hợp Đặc biệt

#### Trường hợp 1: Chuyển giao Ngược (Từ Tổ chức về Cá nhân)

**Khi nào xảy ra:**

- Hợp đồng kết thúc, nhà xuất bản trả quyền cho tác giả
- Tác giả mua lại quyền sở hữu
- Thỏa thuận chấm dứt hợp tác

**Quy trình:**

1. Quản trị viên Tổ chức khởi tạo chuyển giao ngược
2. Chọn tác giả đích (phải là tác giả gốc hoặc người được ủy quyền)
3. Tác giả xác nhận đồng ý nhận lại
4. Hệ thống chuyển đổi: Tổ chức → Cá nhân
5. Lưu lịch sử chuyển giao

#### Trường hợp 2: Chuyển giao giữa 2 Tổ chức

**Khi nào xảy ra:**

- Nhà xuất bản A bán quyền cho nhà xuất bản B
- Sáp nhập công ty

**Quy trình:**

1. Quản trị viên Tổ chức A khởi tạo
2. Chọn Tổ chức B đích
3. Quản trị viên Tổ chức B phê duyệt
4. Hệ thống chuyển đổi: Tổ chức A → Tổ chức B
5. Tác giả gốc nhận thông báo (để biết)

#### Trường hợp 3: Hủy Chuyển giao

**Khi nào có thể hủy:**

- Yêu cầu đang ở trạng thái **Chờ phê duyệt**
- Chưa được phê duyệt

**Ai có thể hủy:**

- Người khởi tạo (tác giả hoặc quản trị viên)

**Hành động:**

- Đặt trạng thái: **Đã hủy**
- Không thay đổi quyền sở hữu
- Lưu lịch sử (để tra cứu)

---

## 7. Phân quyền và Truy cập

### 7.1 Ma trận Phân quyền Chi tiết

#### Với Truyện Sở hữu Cá nhân

| Hành động    | Chủ sở hữu (Tác giả) | Người khác                   | Quản trị viên hệ thống |
| ------------ | -------------------- | ---------------------------- | ---------------------- |
| Xem truyện   | ✅ Luôn luôn         | ❌ Không (trừ khi công khai) | ✅ Có (để kiểm duyệt)  |
| Sửa nội dung | ✅ Có                | ❌ Không                     | ❌ Không               |
| Xóa truyện   | ✅ Có                | ❌ Không                     | ⚠️ Có (nếu vi phạm)    |
| Chuyển giao  | ✅ Có                | ❌ Không                     | ❌ Không               |
| Xuất bản     | ✅ Có                | ❌ Không                     | ❌ Không               |

#### Với Truyện Sở hữu Tổ chức

| Hành động    | Tác giả gốc              | Thành viên Tổ chức       | Quản trị viên Tổ chức | Quản trị viên hệ thống |
| ------------ | ------------------------ | ------------------------ | --------------------- | ---------------------- |
| Xem truyện   | ⚠️ Nếu là thành viên     | ✅ Có                    | ✅ Có                 | ✅ Có                  |
| Sửa nội dung | ⚠️ Nếu có quyền biên tập | ⚠️ Nếu có quyền biên tập | ✅ Có                 | ❌ Không               |
| Xóa truyện   | ❌ Không                 | ❌ Không                 | ✅ Có                 | ⚠️ Có (nếu vi phạm)    |
| Chuyển giao  | ❌ Không                 | ❌ Không                 | ✅ Có                 | ❌ Không               |
| Xuất bản     | ❌ Không                 | ⚠️ Nếu có quyền xuất bản | ✅ Có                 | ❌ Không               |

#### Với Truyện Sở hữu Hợp tác

| Hành động    | Tác giả gốc         | Thành viên Tổ chức       | Quản trị viên Tổ chức | Quản trị viên hệ thống |
| ------------ | ------------------- | ------------------------ | --------------------- | ---------------------- |
| Xem truyện   | ✅ Có               | ✅ Có                    | ✅ Có                 | ✅ Có                  |
| Sửa nội dung | ✅ Có               | ⚠️ Nếu có quyền biên tập | ✅ Có                 | ❌ Không               |
| Xóa truyện   | ⚠️ Cần đồng ý 2 bên | ❌ Không                 | ⚠️ Cần đồng ý 2 bên   | ⚠️ Có (nếu vi phạm)    |
| Chuyển giao  | ⚠️ Cần đồng ý 2 bên | ❌ Không                 | ⚠️ Cần đồng ý 2 bên   | ❌ Không               |
| Xuất bản     | ✅ Có               | ⚠️ Nếu có quyền xuất bản | ✅ Có                 | ❌ Không               |

**Chú thích:**

- ✅ Có quyền vô điều kiện
- ❌ Không có quyền
- ⚠️ Có điều kiện (xem chi tiết trong ô)

### 7.2 Vai trò trong Tổ chức

Khi một người dùng gia nhập Tổ chức, họ được gán một vai trò. Vai trò quyết định quyền hạn.

| Vai trò           | Xem truyện | Sửa truyện         | Xóa truyện                      | Quản lý thành viên | Phê duyệt chuyển giao |
| ----------------- | ---------- | ------------------ | ------------------------------- | ------------------ | --------------------- |
| **Người xem**     | ✅         | ❌                 | ❌                              | ❌                 | ❌                    |
| **Tác giả**       | ✅         | ⚠️ Truyện của mình | ❌                              | ❌                 | ❌                    |
| **Biên tập viên** | ✅         | ✅                 | ❌                              | ❌                 | ❌                    |
| **Quản lý**       | ✅         | ✅                 | ⚠️ Không phải truyện quan trọng | ✅                 | ✅                    |
| **Quản trị viên** | ✅         | ✅                 | ✅                              | ✅                 | ✅                    |

### 7.3 Cấp độ Truy cập

Mỗi truyện có một "Cấp độ Truy cập" quy định ai được xem:

| Cấp độ             | Mô tả                  | Ai xem được?                             |
| ------------------ | ---------------------- | ---------------------------------------- |
| **Riêng tư**       | Chỉ chủ sở hữu         | Chủ sở hữu + Quản trị viên hệ thống      |
| **Nội bộ Tổ chức** | Chỉ thành viên Tổ chức | Tất cả thành viên Tổ chức (theo vai trò) |
| **Công khai**      | Mọi người              | Tất cả người dùng hệ thống               |

**Ví dụ:**

- Truyện Cá nhân, Riêng tư → Chỉ tác giả xem
- Truyện Tổ chức, Nội bộ → Chỉ thành viên nhà xuất bản xem
- Truyện Tổ chức, Công khai → Mọi người xem, nhưng chỉ nhà xuất bản sửa

---

## 8. Câu hỏi Thường gặp

### 8.1 Về Quyền Sở hữu

**Q1: Tôi có thể chuyển truyện cho bất kỳ Tổ chức nào không?**

A: Không. Bạn chỉ có thể chuyển truyện cho Tổ chức mà bạn đang là thành viên. Lý do:

- Đảm bảo có mối quan hệ pháp lý (hợp đồng lao động, hợp tác)
- Tránh spam hoặc chuyển giao nhầm

**Q2: Sau khi chuyển giao hoàn toàn, tôi còn quyền gì với truyện?**

A: Bạn vẫn được ghi nhận là "Tác giả gốc", nhưng mất quyền kiểm soát:

- ❌ Không thể sửa nội dung (trừ khi Tổ chức cấp quyền biên tập viên)
- ❌ Không thể xóa truyện
- ✅ Tên bạn vẫn hiển thị là tác giả
- ✅ Bạn có thể yêu cầu Tổ chức chuyển lại (nếu thỏa thuận)

**Q3: Nếu tôi rời Tổ chức, truyện tôi đã chuyển giao có theo tôi không?**

A: Tùy loại chuyển giao:

- **Chuyển giao hoàn toàn:** Truyện ở lại với Tổ chức
- **Hợp tác chia sẻ:** Bạn có thể yêu cầu chuyển về Cá nhân (cần phê duyệt)

**Q4: Tôi có thể hủy chuyển giao không?**

A: Chỉ khi yêu cầu đang ở trạng thái "Chờ phê duyệt". Sau khi phê duyệt, phải thỏa thuận với Tổ chức để chuyển giao ngược.

### 8.2 Về Phân quyền

**Q5: Tôi là thành viên của 3 Tổ chức, làm sao biết đang làm việc cho Tổ chức nào?**

A: Hệ thống có thanh chọn "Ngữ cảnh Tổ chức" ở góc trên:

- Chọn "Cá nhân" → Truyện mới sẽ thuộc về bạn
- Chọn "Tổ chức A" → Truyện mới thuộc về Tổ chức A
- Luôn kiểm tra trước khi tạo truyện mới

**Q6: Nếu tôi vừa là tác giả gốc, vừa là biên tập viên của Tổ chức, quyền của tôi thế nào?**

A: Bạn có quyền cao nhất trong 2 vai trò:

- Với truyện Cá nhân của bạn → Quyền chủ sở hữu đầy đủ
- Với truyện Tổ chức (của người khác) → Quyền biên tập viên
- Với truyện Tổ chức (bạn là tác giả gốc) → Quyền biên tập viên (không có quyền đặc biệt hơn)

**Q7: Quản trị viên hệ thống có thể làm gì với truyện của tôi?**

A: Quản trị viên hệ thống (WibuSystem) chỉ can thiệp khi:

- ⚠️ Truyện vi phạm quy định (nội dung bất hợp pháp, bản quyền...)
- ⚠️ Có khiếu nại từ người dùng
- Trong trường hợp bình thường, họ KHÔNG sửa hoặc xóa truyện của bạn

### 8.3 Về Quy trình

**Q8: Mất bao lâu để một yêu cầu chuyển giao được phê duyệt?**

A: Tùy thuộc vào Tổ chức:

- Quản trị viên Tổ chức quyết định thời gian xử lý
- Hệ thống sẽ gửi thông báo nhắc nhở sau 7 ngày nếu chưa xử lý
- Bạn có thể hủy yêu cầu và tạo lại nếu quá lâu

**Q9: Tôi có thể chuyển nhiều truyện cùng lúc không?**

A: Có. Khi tạo yêu cầu chuyển giao, bạn có thể:

- Chọn nhiều truyện (tối đa 50 truyện/lần)
- Áp dụng cùng một loại chuyển giao cho tất cả

**Q10: Nếu Tổ chức từ chối yêu cầu chuyển giao, điều gì xảy ra?**

A: Truyện vẫn thuộc về bạn, không thay đổi gì. Bạn có thể:

- Xem lý do từ chối
- Trao đổi với Tổ chức
- Tạo yêu cầu mới với thông tin rõ ràng hơn

### 8.4 Về Đặc thù Nghiệp vụ

**Q11: Mô hình "Hợp tác chia sẻ" khác gì "Chuyển giao hoàn toàn"?**

A:

| Tiêu chí          | Chuyển giao hoàn toàn | Hợp tác chia sẻ                    |
| ----------------- | --------------------- | ---------------------------------- |
| Chủ sở hữu        | Tổ chức               | Tác giả (nhưng chia sẻ quyền)      |
| Tác giả sửa được? | ❌ Không              | ✅ Có                              |
| Tác giả xóa được? | ❌ Không              | ⚠️ Cần đồng ý 2 bên                |
| Phù hợp khi nào?  | Hợp đồng độc quyền    | Chia sẻ doanh thu, hợp tác lâu dài |

**Q12: Nếu Tổ chức giải thể, truyện sẽ thế nào?**

A: Hệ thống có quy trình xử lý:

1. Quản trị viên hệ thống thông báo giải thể
2. Tất cả truyện Sở hữu Tổ chức được chuyển về tác giả gốc (tự động)
3. Nếu không tìm được tác giả gốc → Truyện được lưu trữ an toàn, chờ yêu cầu

**Q13: Tôi có thể chuyển truyện từ Tổ chức A sang Tổ chức B không?**

A: Không trực tiếp. Quy trình:

1. Yêu cầu Tổ chức A chuyển về Cá nhân (chuyển giao ngược)
2. Sau khi về Cá nhân, bạn chuyển sang Tổ chức B

Hoặc: Quản trị viên Tổ chức A có thể chuyển trực tiếp sang Tổ chức B (chuyển giao giữa 2 Tổ chức).

---

## 9. Lộ trình Triển khai

### 9.1 Giai đoạn 1: Thiết kế và Chuẩn bị (Tuần 1-2)

**Mục tiêu:** Hoàn thiện tài liệu và thiết kế giao diện

**Công việc:**

| Công việc                        | Người phụ trách              | Kết quả mong đợi                            |
| -------------------------------- | ---------------------------- | ------------------------------------------- |
| Duyệt tài liệu nghiệp vụ này     | Quản lý Sản phẩm, Kinh doanh | Tài liệu được phê duyệt                     |
| Thiết kế giao diện (UI/UX)       | Designer                     | Mockup màn hình chuyển giao, quản lý truyện |
| Lập kế hoạch kỹ thuật            | Tech Lead                    | Tài liệu kỹ thuật chi tiết                  |
| Ước lượng thời gian và nguồn lực | PM + Tech Lead               | Kế hoạch triển khai chi tiết                |

**Deliverables:**

- ✅ Tài liệu nghiệp vụ phê duyệt
- ✅ Thiết kế UI/UX
- ✅ Tài liệu kỹ thuật
- ✅ Kế hoạch chi tiết

### 9.2 Giai đoạn 2: Phát triển Nền tảng (Tuần 3-5)

**Mục tiêu:** Xây dựng các tính năng cốt lõi

**Công việc:**

| Tính năng                           | Mô tả                              | Ưu tiên        |
| ----------------------------------- | ---------------------------------- | -------------- |
| Hệ thống quyền sở hữu               | Lưu trữ 3 loại sở hữu, tác giả gốc | P0 - Bắt buộc  |
| Chuyển giao Cá nhân → Tổ chức       | Tạo yêu cầu, phê duyệt, thực hiện  | P0 - Bắt buộc  |
| Phân quyền theo vai trò             | Kiểm tra quyền xem/sửa/xóa         | P0 - Bắt buộc  |
| Lịch sử chuyển giao                 | Lưu và hiển thị lịch sử            | P1 - Nên có    |
| Chuyển giao ngược Tổ chức → Cá nhân | Tính năng chuyển lại               | P1 - Nên có    |
| Hợp tác chia sẻ                     | Quản lý đồng sở hữu                | P2 - Có thể có |

**Deliverables:**

- ✅ Tính năng P0 hoàn thành và test
- ✅ Tính năng P1 hoàn thành cơ bản

### 9.3 Giai đoạn 3: Giao diện Người dùng (Tuần 6-7)

**Mục tiêu:** Xây dựng giao diện trực quan, dễ sử dụng

**Màn hình cần phát triển:**

1. **Màn hình "Truyện của tôi"**

   - Hiển thị truyện theo nhóm: Cá nhân, Tổ chức A, Tổ chức B...
   - Badge hiển thị loại sở hữu
   - Nút "Chuyển giao" cho từng truyện

2. **Màn hình "Chuyển giao truyện"**

   - Chọn truyện (checkbox nhiều truyện)
   - Chọn Tổ chức đích
   - Chọn loại chuyển giao
   - Nhập lý do
   - Xem trước kết quả

3. **Màn hình "Lịch sử chuyển giao"**

   - Danh sách các lần chuyển giao
   - Lọc theo trạng thái, ngày tháng
   - Xem chi tiết từng lần

4. **Màn hình "Phê duyệt chuyển giao"** (Cho quản trị viên Tổ chức)
   - Danh sách yêu cầu chờ phê duyệt
   - Chi tiết truyện muốn nhận
   - Nút Phê duyệt / Từ chối

**Deliverables:**

- ✅ Tất cả màn hình hoàn thành
- ✅ Responsive (mobile, tablet, desktop)

### 9.4 Giai đoạn 4: Kiểm thử (Tuần 8)

**Mục tiêu:** Đảm bảo chất lượng trước khi ra mắt

**Kịch bản kiểm thử:**

| Kịch bản                        | Kết quả mong đợi                                |
| ------------------------------- | ----------------------------------------------- |
| Tác giả tạo truyện cá nhân      | Truyện có Sở hữu Cá nhân, chỉ tác giả xem được  |
| Tác giả chuyển giao cho Tổ chức | Yêu cầu được tạo, quản trị viên nhận thông báo  |
| Quản trị viên phê duyệt         | Quyền sở hữu thay đổi đúng, lịch sử được lưu    |
| Quản trị viên từ chối           | Truyện vẫn ở dạng cũ, tác giả nhận thông báo    |
| Tác giả rời Tổ chức             | Hệ thống xử lý đúng với từng loại truyện        |
| Tác giả thuộc nhiều Tổ chức     | Chọn ngữ cảnh đúng, tạo truyện cho đúng Tổ chức |

**Kiểm thử hiệu năng:**

- Hệ thống xử lý 1000 yêu cầu chuyển giao đồng thời
- Hiển thị 10,000 truyện trong "Truyện của tôi"
- Tìm kiếm truyện trong 1 triệu bản ghi

**Deliverables:**

- ✅ Báo cáo kiểm thử
- ✅ Tất cả lỗi P0, P1 được sửa

### 9.5 Giai đoạn 5: Triển khai Thử nghiệm (Tuần 9-10)

**Mục tiêu:** Cho một nhóm nhỏ người dùng dùng thử

**Nhóm thử nghiệm:**

- 10 tác giả cá nhân
- 3 nhà xuất bản nhỏ (mỗi nhà 5-10 thành viên)

**Quy trình:**

1. Mời nhóm thử nghiệm
2. Cung cấp hướng dẫn sử dụng
3. Thu thập phản hồi qua:
   - Khảo sát trực tuyến
   - Phỏng vấn 1-1
   - Phân tích hành vi sử dụng
4. Điều chỉnh dựa trên phản hồi

**Deliverables:**

- ✅ Báo cáo phản hồi người dùng
- ✅ Danh sách cải tiến cần làm

### 9.6 Giai đoạn 6: Ra mắt Chính thức (Tuần 11-12)

**Mục tiêu:** Triển khai cho toàn bộ người dùng

**Kế hoạch ra mắt:**

**Tuần 11:**

- Thứ 2: Triển khai lên Production
- Thứ 3: Thông báo cho tất cả người dùng qua email
- Thứ 4: Đăng bài blog giới thiệu tính năng
- Thứ 5: Webinar hướng dẫn sử dụng
- Thứ 6: Theo dõi lỗi, sửa nhanh

**Tuần 12:**

- Hỗ trợ người dùng qua chat/email
- Thu thập phản hồi
- Lập kế hoạch cải tiến v2

**Deliverables:**

- ✅ Tính năng hoạt động ổn định
- ✅ Tài liệu hướng dẫn người dùng
- ✅ Video tutorial
- ✅ Đội ngũ hỗ trợ sẵn sàng

---

## 10. Phụ lục

### 10.1 Thuật ngữ

| Thuật ngữ                 | Định nghĩa                                                               |
| ------------------------- | ------------------------------------------------------------------------ |
| **Tác giả gốc**           | Người đầu tiên tạo ra tác phẩm, không thay đổi dù quyền sở hữu chuyển đi |
| **Chủ sở hữu**            | Cá nhân hoặc Tổ chức có quyền kiểm soát chính đối với tác phẩm           |
| **Tổ chức (Tenant)**      | Nhà xuất bản, công ty, hoặc nhóm quản lý nội dung                        |
| **Thành viên**            | Người dùng gia nhập một Tổ chức, có vai trò cụ thể                       |
| **Chuyển giao**           | Hành động thay đổi quyền sở hữu từ bên này sang bên khác                 |
| **Chuyển giao hoàn toàn** | Chuyển giao toàn bộ quyền kiểm soát, bên cũ mất quyền                    |
| **Hợp tác chia sẻ**       | Cả hai bên cùng có quyền quản lý tác phẩm                                |
| **Ngữ cảnh Tổ chức**      | Tổ chức đang được chọn để làm việc trong phiên hiện tại                  |
| **Phê duyệt**             | Quá trình xem xét và chấp nhận yêu cầu chuyển giao                       |

### 10.2 Ví dụ Thực tế

#### Ví dụ 1: Tác giả Web Novel

**Bối cảnh:**

- Nguyễn Văn A viết web novel "Kiếm Hiệp Giang Hồ" trên WibuSystem
- Sau 6 tháng, truyện có 100,000 lượt xem
- Nhà xuất bản "Kim Dung Books" mời hợp tác

**Hành động:**

1. A gia nhập "Kim Dung Books" với vai trò "Tác giả"
2. A chuyển giao "Kiếm Hiệp Giang Hồ" cho Kim Dung Books (Chuyển giao hoàn toàn)
3. Kim Dung Books biên tập, đóng bìa, xuất bản
4. Doanh thu chia 70% (Kim Dung) - 30% (A) theo hợp đồng

**Kết quả:**

- Truyện thuộc quyền sở hữu của Kim Dung Books
- Tên A vẫn hiện là "Tác giả: Nguyễn Văn A"
- A không tự sửa truyện (phải qua biên tập viên Kim Dung)
- Mọi quyết định xuất bản do Kim Dung quyết định

#### Ví dụ 2: Tác giả Đa nền tảng

**Bối cảnh:**

- Trần Thị B là tác giả có tài khoản trên nhiều nền tảng
- B có 5 truyện cá nhân trên WibuSystem
- B vừa ký hợp đồng với "Nhà xuất bản Trẻ" và "Nhà xuất bản Kim Đồng"

**Hành động:**

1. B gia nhập cả 2 Tổ chức
2. B chuyển:
   - Truyện "Mắt Biếc" → Nhà xuất bản Trẻ (Chuyển giao hoàn toàn)
   - Truyện "Cho Tôi Xin Một Vé Đi Tuổi Thơ" → Nhà xuất bản Trẻ (Hợp tác)
   - Truyện "Tôi Thấy Hoa Vàng" → Nhà xuất bản Kim Đồng (Chuyển giao hoàn toàn)
   - Giữ lại 2 truyện fanfiction cá nhân

**Kết quả:**

- B thấy truyện của mình trong 3 nhóm:
  - Cá nhân: 2 truyện
  - Nhà xuất bản Trẻ: 2 truyện
  - Nhà xuất bản Kim Đồng: 1 truyện
- Khi tạo truyện mới, B chọn "Tạo cho Tổ chức nào?"

### 10.3 Checklist Triển khai

**Trước khi ra mắt:**

- [ ] Tất cả tính năng P0 hoàn thành
- [ ] Kiểm thử kỹ lưỡng (Unit, Integration, E2E)
- [ ] Tài liệu người dùng hoàn thành
- [ ] Video hướng dẫn quay xong
- [ ] Đội ngũ hỗ trợ được đào tạo
- [ ] Thông báo email soạn sẵn
- [ ] Kế hoạch rollback (nếu có lỗi nghiêm trọng)

**Ngày ra mắt:**

- [ ] Triển khai lên Production vào sáng sớm
- [ ] Kiểm tra lại tất cả tính năng
- [ ] Gửi thông báo email
- [ ] Đăng bài blog/social media
- [ ] Theo dõi lỗi real-time
- [ ] Chuẩn bị hotfix nếu cần

**Sau ra mắt:**

- [ ] Thu thập phản hồi người dùng
- [ ] Phân tích số liệu sử dụng
- [ ] Sửa lỗi phát sinh
- [ ] Lập kế hoạch cải tiến
- [ ] Báo cáo kết quả cho ban lãnh đạo

---
