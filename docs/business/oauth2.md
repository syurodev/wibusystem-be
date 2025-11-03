# **Kiến trúc Máy chủ OAuth 2.0 Đa Khách hàng, Tuân thủ Tiêu chuẩn với Go, Fosite và PostgreSQL**

## **Bản thiết kế Kiến trúc cho một Dịch vụ Phân quyền Hiện đại**

Tài liệu này phác thảo thiết kế kiến trúc cho một máy chủ phân quyền OAuth 2.0 tách biệt, tuân thủ tiêu chuẩn và có khả năng mở rộng cao. Nguyên tắc thiết kế cốt lõi là tách biệt các mối quan tâm, cụ thể là phân biệt giữa xác thực (xác minh danh tính người dùng) và phân quyền (xác định những gì người dùng đã xác thực được phép làm). Hệ thống này được thiết kế để hoạt động như một cơ quan phân quyền trung tâm, trong khi logic xác thực (ví dụ: biểu mẫu đăng nhập, quản lý mật khẩu) vẫn là trách nhiệm của các ứng dụng tích hợp. Bằng cách tận dụng một ngăn xếp công nghệ hiện đại, kiến trúc này ưu tiên bảo mật, hiệu năng và khả năng bảo trì cho môi trường SaaS đa khách hàng (multi-tenant).

### **Tổng quan Hệ thống và Vai trò của các Thành phần**

Hệ thống bao gồm một số thành phần chính, mỗi thành phần được lựa chọn vì những thế mạnh cụ thể trong việc xây dựng một dịch vụ phân quyền mạnh mẽ và hiệu năng cao. Sự tương tác giữa các thành phần này tạo thành một kiến trúc gắn kết và có khả năng mở rộng.

* **Go với Gin Framework**: Lớp HTTP của ứng dụng được xây dựng bằng ngôn ngữ lập trình Go và framework web Gin. Go cung cấp hiệu năng cao, các cơ chế đồng thời mạnh mẽ và một thư viện chuẩn vững chắc, làm cho nó trở nên lý tưởng cho các dịch vụ mạng. Gin là một framework web tối giản và nhanh chóng, cung cấp một router dựa trên cây radix hiệu suất cao và một API tiện lợi để xây dựng các dịch vụ RESTful mà không có chi phí không cần thiết.1 Trong kiến trúc này, vai trò của Gin là phơi bày các endpoint OAuth 2.0 tiêu chuẩn—/authorize, /token, /introspect, và /revoke—và điều phối luồng yêu cầu đến lõi phân quyền.2
* **Thư viện ORY Fosite**: Trung tâm của hệ thống là ORY Fosite, một SDK Go ưu tiên bảo mật để triển khai các nhà cung cấp OAuth 2.0 và OpenID Connect.3 Fosite không phải là một máy chủ độc lập mà là một framework cung cấp logic để xử lý sự phức tạp của đặc tả OAuth 2.0, bao gồm các loại grant khác nhau, tạo và xác thực token.5 Cách tiếp cận dựa trên SDK này cho phép kiểm soát hoàn toàn việc triển khai, đặc biệt là backend lưu trữ và luồng chấp thuận của người dùng, cho phép một giải pháp tùy chỉnh sâu và tích hợp tuân thủ nghiêm ngặt RFC 6749 và các tiêu chuẩn liên quan.4
* **PostgreSQL 18**: PostgreSQL đóng vai trò là lớp lưu trữ chính cho tất cả dữ liệu bền vững, có quan hệ. Nó là hệ thống ghi nhận chính thức cho thông tin đòi hỏi tính toàn vẹn giao dịch, tính nhất quán mạnh mẽ và lưu trữ lâu dài. Điều này bao gồm dữ liệu nền tảng như danh tính người dùng, hồ sơ khách hàng (tenant), cấu hình client và cấu trúc Kiểm soát Truy cập Dựa trên Vai trò (RBAC) toàn diện gồm các vai trò và quyền.7 Việc lựa chọn PostgreSQL 18 là có chủ ý, vì nó giới thiệu hỗ trợ gốc cho UUID v7, một tính năng quan trọng để tối ưu hóa hiệu năng cơ sở dữ liệu trong kiến trúc này.8
* **Redis**: Redis được sử dụng như một kho dữ liệu trong bộ nhớ, tốc độ cao cho dữ liệu tạm thời và được truy cập thường xuyên. Chức năng chính của nó là giảm tải các hoạt động có thông lượng cao từ PostgreSQL, do đó nâng cao hiệu năng và khả năng phản hồi tổng thể của hệ thống. Redis sẽ quản lý các tạo tác OAuth 2.0 tồn tại trong thời gian ngắn như mã ủy quyền (authorization code), các phiên PKCE và chữ ký của các token đang hoạt động. Nó cũng sẽ đóng vai trò là một danh sách đen tập trung để thực hiện việc thu hồi token ngay lập tức, một tính năng bảo mật quan trọng.9

### **Chiến lược Lưu trữ Lai: Kết hợp PostgreSQL và Redis**

Một chiến lược lưu trữ lai là nền tảng để đạt được các mức hiệu năng, khả năng mở rộng và khả năng phục hồi cần thiết. Chỉ dựa vào một cơ sở dữ liệu quan hệ như PostgreSQL cho mọi hoạt động—từ việc xác thực thông tin đăng nhập của client đến việc kiểm tra (introspect) một access token trên mỗi lệnh gọi API—sẽ tạo ra một điểm nghẽn hiệu năng đáng kể và hạn chế khả năng mở rộng của hệ thống dưới tải nặng.

* **Vai trò của PostgreSQL (Hệ thống Ghi nhận Chính thức)**: PostgreSQL được chỉ định là nguồn dữ liệu chính xác và có thẩm quyền cho dữ liệu bền vững và đòi hỏi tính toàn vẹn quan hệ. Nó lưu trữ thông tin ít thay đổi nhưng phải được lưu trữ một cách đáng tin cậy. Điều này bao gồm:
    * Tài khoản và hồ sơ người dùng.
    * Bản ghi khách hàng (tổ chức).
    * Cấu hình client OAuth 2.0 (oauth2_clients).
    * Lược đồ RBAC hoàn chỉnh: roles, permissions, và các mối quan hệ của chúng.
      Bằng cách giới hạn vai trò của PostgreSQL trong việc quản lý dữ liệu nền tảng này, khối lượng công việc của nó được tối ưu hóa cho tính nhất quán giao dịch thay vì các thao tác đọc và ghi tần suất cao, độ trễ thấp của dữ liệu tạm thời.
* **Vai trò của Redis (Bộ đệm Tốc độ cao & Kho lưu trữ Tạm thời)**: Redis được tối ưu hóa để xử lý dữ liệu tồn tại trong thời gian ngắn hoặc yêu cầu thời gian truy cập dưới một mili giây.9 Các trách nhiệm của nó bao gồm:
    * **Lưu trữ Mã ủy quyền (Authorization Codes)**: Các mã này chỉ được sử dụng một lần và có tuổi thọ rất ngắn (thường là 60 giây), làm cho chúng trở thành ứng cử viên lý tưởng để lưu trữ trong Redis với cài đặt tự động hết hạn (EX).
    * **Quản lý các Phiên PKCE**: Tương tự như mã ủy quyền, dữ liệu thử thách Proof Key for Code Exchange (PKCE) là tạm thời và chỉ cần thiết trong suốt một lần trao đổi token.
    * **Lưu trữ Token đang hoạt động**: Lưu trữ chữ ký hoặc định danh của các access token và refresh token đang hoạt động cho phép kiểm tra nhanh chóng mà không cần truy vấn cơ sở dữ liệu chính.
    * **Danh sách Thu hồi Token**: Một tập hợp Redis hoặc kho lưu trữ khóa-giá trị cung cấp một cơ chế cực kỳ nhanh để kiểm tra xem một token đã bị thu hồi rõ ràng trước khi hết hạn tự nhiên hay chưa.12

Sự phân chia công việc này đảm bảo rằng khối lượng lớn các hoạt động liên quan đến việc cấp và xác thực token được xử lý bởi thành phần phù hợp nhất cho nhiệm vụ—Redis—trong khi tính toàn vẹn của dữ liệu cấu hình và nhận dạng cốt lõi được đảm bảo bởi PostgreSQL.

### **Sơ đồ Kiến trúc Cấp cao**

Sơ đồ sau đây minh họa sự tương tác giữa các thành phần của hệ thống trong một luồng phân quyền điển hình. Nó làm nổi bật sự tách biệt giữa các đường giao tiếp kênh phía trước (chuyển hướng dựa trên trình duyệt) và kênh phía sau (lệnh gọi API từ máy chủ đến máy chủ).

Đoạn mã

sequenceDiagram
participant UserAgent as User-Agent (Trình duyệt)
participant ClientApp as Ứng dụng Client
participant AuthServer as Máy chủ Gin/Fosite
participant Redis
participant PostgreSQL

    UserAgent->>+ClientApp: 1. Người dùng bắt đầu đăng nhập
    ClientApp->>UserAgent: 2. Chuyển hướng đến endpoint /authorize
    UserAgent->>+AuthServer: 3. GET /authorize (với client_id, scope, v.v.)
    AuthServer-->>UserAgent: 4. Hiển thị Giao diện Đăng nhập & Chấp thuận
    UserAgent->>AuthServer: 5. Người dùng gửi thông tin xác thực & chấp thuận
    AuthServer->>PostgreSQL: 6. Xác minh thông tin người dùng
    AuthServer->>Redis: 7. Lưu Authorization Code & session PKCE
    AuthServer-->>-UserAgent: 8. Chuyển hướng đến ClientApp với `code`
    UserAgent->>ClientApp: 9. Gửi `code` đến client
    ClientApp->>+AuthServer: 10. POST /token (kênh sau) với `code` & `client_secret`
    AuthServer->>Redis: 11. Xác thực `code`, lấy và xóa session
    AuthServer->>PostgreSQL: 12. Xác thực `client_secret`
    AuthServer->>Redis: 13. Lưu chữ ký token mới để kiểm tra
    AuthServer-->>-ClientApp: 14. Trả về `access_token` & `refresh_token`
    ClientApp->>AuthServer: 15. Gọi Resource Server (không hiển thị) với `access_token`

Cách tiếp cận kiến trúc này đảm bảo sự tách biệt rõ ràng các mối quan tâm. Framework Gin hoạt động như điểm vào và người điều phối, nhưng nó vẫn là một lớp mỏng. Logic cốt lõi được đóng gói trong Fosite provider, được cấu hình để sử dụng một backend lưu trữ thông minh ủy quyền các hoạt động cho PostgreSQL hoặc Redis. Thiết kế này không chỉ hiệu năng mà còn có tính mô-đun cao. Ví dụ, framework web có thể được thay đổi từ Gin sang một framework khác, như net/http, với những thay đổi tối thiểu đối với logic phân quyền và lưu trữ cơ bản, thể hiện một thế mạnh chính của kiến trúc tách biệt này.

## **Thiết kế Lớp Lưu trữ RBAC Đa Khách hàng trên PostgreSQL 18**

Nền tảng của bất kỳ hệ thống phân quyền mạnh mẽ nào là mô hình dữ liệu của nó. Phần này trình bày chi tiết về thiết kế lược đồ PostgreSQL, được kiến trúc để hỗ trợ một môi trường đa khách hàng phức tạp với mô hình RBAC hai phạm vi, đồng thời đáp ứng các yêu cầu lưu trữ của thư viện Fosite. Việc lựa chọn UUID v7 làm chiến lược khóa chính là một nền tảng của thiết kế này, nhằm đảm bảo hiệu năng và khả năng mở rộng cao.

### **Mô hình Đa khách hàng: Cơ sở dữ liệu Chung, Lược đồ Chung**

Khi thiết kế một ứng dụng đa khách hàng, việc lựa chọn mô hình đa khách hàng có ảnh hưởng sâu sắc đến khả năng mở rộng, sự cô lập dữ liệu, chi phí và sự phức tạp trong vận hành. Các mô hình chính là Cơ sở dữ liệu cho mỗi Khách hàng (Database-per-Tenant), Lược đồ cho mỗi Khách hàng (Schema-per-Tenant), và Cơ sở dữ liệu/Lược đồ Chung (Shared Database/Shared Schema).14
Đối với hệ thống này, mô hình **Cơ sở dữ liệu Chung, Lược đồ Chung** là lựa chọn phù hợp nhất. Trong mô hình này, dữ liệu của tất cả các khách hàng cùng tồn tại trong cùng một cơ sở dữ liệu và các bảng. Sự cô lập dữ liệu được thực hiện bằng cách thêm một cột tenant_id vào mỗi bảng chứa dữ liệu dành riêng cho khách hàng. Mọi truy vấn cơ sở dữ liệu sau đó phải được giới hạn phạm vi bằng một mệnh đề WHERE tenant_id =? để đảm bảo rằng một khách hàng không bao giờ có thể truy cập dữ liệu của khách hàng khác.15
Mô hình này được chọn vì một số lý do thuyết phục:

* **Khả năng mở rộng**: Nó có thể mở rộng đến hàng chục nghìn khách hàng hiệu quả hơn nhiều so với các mô hình khác, vốn yêu cầu cung cấp cơ sở dữ liệu hoặc lược đồ mới cho mỗi khách hàng mới, tạo ra chi phí vận hành đáng kể.15
* **Hiệu quả về chi phí**: Chia sẻ một phiên bản cơ sở dữ liệu duy nhất hiệu quả hơn đáng kể về tài nguyên và chi phí so với việc quản lý hàng nghìn cơ sở dữ liệu riêng biệt.16
* **Đơn giản trong vận hành**: Việc di chuyển lược đồ (schema migration), sao lưu và giám sát được đơn giản hóa vì chúng áp dụng cho một cơ sở dữ liệu duy nhất thay vì một loạt cơ sở dữ liệu.

Thách thức chính của mô hình này—thực thi sự cô lập dữ liệu ở cấp ứng dụng—là một vấn đề có thể quản lý được và đã được hiểu rõ, sẽ được xử lý nghiêm ngặt trong lớp truy cập dữ liệu của ứng dụng Go.

### **Lược đồ RBAC Đa Khách hàng và Nhận dạng Cốt lõi**

Lược đồ sau đây mở rộng mô hình RBAC tiêu chuẩn để hỗ trợ cả đa khách hàng và sự tách biệt cần thiết giữa các phạm vi "toàn cục" (global) và "khách hàng" (tenant) cho các vai trò và quyền.17 Tất cả các khóa chính đều có kiểu UUID và sẽ mặc định là uuidv7().

* **tenants**: Lưu trữ thông tin về mỗi tổ chức khách hàng.
```sql
CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  name VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'active', -- ví dụ: active, suspended
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
```

* **users**: Một bảng toàn cục cho tất cả các tài khoản người dùng. Một người dùng có thể tồn tại độc lập với bất kỳ khách hàng nào.
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  -- Các trường hồ sơ khác như name, avatar_url, v.v.
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
```

* **permissions**: Một danh sách chính của tất cả các hành động có thể có trong hệ thống, được phân loại theo phạm vi.
```sql
CREATE TYPE permission_scope AS ENUM ('global', 'tenant');

  CREATE TABLE permissions (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  name VARCHAR(255) UNIQUE NOT NULL, -- ví dụ: 'user:view_self', 'content:create_anime'
  scope permission_scope NOT NULL,
  description TEXT
  );
```

* **roles**: Một danh sách chính của tất cả các vai trò có thể có, cũng được phân loại theo phạm vi.
```sql
CREATE TYPE role_scope AS ENUM ('global', 'tenant');

  CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  name VARCHAR(255) UNIQUE NOT NULL, -- ví dụ: 'SUPER_ADMIN', 'TENANT_ADMIN'
  scope role_scope NOT NULL,
  description TEXT
  );
```

* **role_permissions**: Một bảng nối nhiều-nhiều liên kết các vai trò với các quyền được phép của chúng.
```sql
CREATE TABLE role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);
```

* **user_tenant_memberships**: Liên kết quan trọng kết nối một người dùng với một khách hàng.
```sql
CREATE TABLE user_tenant_memberships (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  status VARCHAR(50) NOT NULL DEFAULT 'active', -- ví dụ: active, pending_invite
  PRIMARY KEY (user_id, tenant_id)
);
```

* **user_tenant_roles**: Gán vai trò cho một người dùng *trong ngữ cảnh của một khách hàng cụ thể*.
```sql
CREATE TABLE user_tenant_roles (
  user_id UUID NOT NULL,
  tenant_id UUID NOT NULL,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, tenant_id, role_id),
  FOREIGN KEY (user_id, tenant_id) REFERENCES user_tenant_memberships(user_id, tenant_id) ON DELETE CASCADE
);
```

* **user_global_roles**: Gán vai trò toàn cục cho một người dùng, độc lập với bất kỳ khách hàng nào.
```sql
CREATE TABLE user_global_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id)
);
```

### **Lược đồ Dành riêng cho Fosite để Lưu trữ Trạng thái OAuth 2.0**

Các bảng này được thiết kế để đáp ứng hợp đồng lưu trữ theo yêu cầu của thư viện Fosite. Thiết kế này được lấy cảm hứng từ kho lưu trữ SQL đã được kiểm chứng qua thực tế từ dự án ORY Hydra, ưu tiên sự linh hoạt và hiệu năng hơn là chuẩn hóa nghiêm ngặt cho dữ liệu tạm thời.18

* **oauth2_clients**: Lưu trữ thông tin về các ứng dụng client OAuth 2.0 đã đăng ký. Một quyết định thiết kế quan trọng ở đây là làm cho tenant_id có thể nhận giá trị NULL. Một giá trị NULL biểu thị một client toàn cục, của bên thứ nhất (ví dụ: một bảng điều khiển quản trị toàn hệ thống), trong khi một giá trị không NULL liên kết client với một khách hàng cụ thể. Điều này cho phép hệ thống phân biệt giữa các ứng dụng nền tảng nội bộ và các ứng dụng bên ngoài hoặc dành riêng cho khách hàng, điều này rất cần thiết để áp dụng các chính sách bảo mật và giới hạn phạm vi khác nhau.
```sql
CREATE TABLE oauth2_clients (
  id UUID PRIMARY KEY DEFAULT uuidv7(),
  client_name VARCHAR(255) NOT NULL,
  secret_hash VARCHAR(255) NOT NULL,
  redirect_uris TEXT NOT NULL,
  grant_types TEXT NOT NULL,
  response_types TEXT NOT NULL,
  scopes TEXT NOT NULL,
  is_public BOOLEAN NOT NULL DEFAULT FALSE,
  tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE, -- Có thể là NULL cho các client toàn cục
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

* **oauth2_sessions**: Đây là một bảng chung được thiết kế để lưu trữ trạng thái được tuần tự hóa của các tạo tác OAuth 2.0 khác nhau. Thay vì tạo các bảng riêng biệt cho mã ủy quyền, access token, refresh token và các phiên PKCE, cách tiếp cận này sử dụng một bảng duy nhất. Mỗi tạo tác được đại diện bởi chữ ký duy nhất của nó, và toàn bộ trạng thái của yêu cầu (fosite.Requester) được lưu trữ dưới dạng một khối JSON. Thiết kế này đơn giản hóa đáng kể lược đồ và làm cho lớp lưu trữ có khả năng chống lại những thay đổi trong cấu trúc dữ liệu nội bộ của Fosite. Nó đánh đổi sự chuẩn hóa cơ sở dữ liệu nghiêm ngặt để lấy được những lợi ích đáng kể về tính linh hoạt và khả năng bảo trì.
```sql
CREATE TYPE session_type AS ENUM ('authorize_code', 'access_token', 'refresh_token', 'pkce', 'openid');

  CREATE TABLE oauth2_sessions (
  signature VARCHAR(255) PRIMARY KEY,
  request_id VARCHAR(255) NOT NULL,
  session_type session_type NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  session_data JSONB NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  client_id UUID NOT NULL REFERENCES oauth2_clients(id) ON DELETE CASCADE,
  subject_id UUID REFERENCES users(id) ON DELETE CASCADE -- Có thể là NULL cho grant client_credentials
  );
  CREATE INDEX idx_oauth2_sessions_request_id ON oauth2_sessions(request_id);
  CREATE INDEX idx_oauth2_sessions_expires_at ON oauth2_sessions(expires_at);
```

* **oauth2_jti_blacklist**: Được sử dụng để ngăn chặn các cuộc tấn công phát lại (replay attack) cho luồng JWT Bearer Grant và có thể được tái sử dụng cho việc thu hồi token nói chung.
```sql
CREATE TABLE oauth2_jti_blacklist (
  signature VARCHAR(255) PRIMARY KEY,
  expires_at TIMESTAMPTZ NOT NULL
  );
CREATE INDEX idx_oauth2_jti_blacklist_expires_at ON oauth2_jti_blacklist(expires_at);
```

### **Chiến lược Khóa chính: Sức mạnh của UUID v7**

Quyết định sử dụng UUID làm khóa chính là tiêu chuẩn cho các hệ thống phân tán, vì chúng có thể được tạo ra một cách độc lập trên bất kỳ nút nào mà không cần phối hợp, không giống như các số nguyên tuần tự.8 Tuy nhiên, việc lựa chọn *phiên bản* UUID là một yếu tố quan trọng về hiệu năng.

| Tính năng | UUID v4 (Ngẫu nhiên) | UUID v7 (Sắp xếp theo thời gian) |
| :---- | :---- | :---- |
| **Tạo ra** | 122 bit dữ liệu ngẫu nhiên an toàn về mặt mật mã. | 48-bit dấu thời gian Unix (độ chính xác ms) + 74 bit dữ liệu ngẫu nhiên. |
| **Khả năng sắp xếp** | Không thể sắp xếp theo thời gian tạo. | Có thể sắp xếp tự nhiên theo thời gian tạo. |
| **Tính cục bộ của chỉ mục** | Kém. Các lần chèn được phân tán ngẫu nhiên trên chỉ mục B-tree. | Tuyệt vời. Các lần chèn là tuần tự, được nối vào cuối chỉ mục. |
| **Hiệu năng ghi** | Giảm dần theo thời gian. Gây ra việc phân chia trang B-tree thường xuyên, dẫn đến chỉ mục bị phình to, tăng I/O và khuếch đại ghi cao hơn.20 | Cao và ổn định. Bắt chước hiệu năng của BIGSERIAL bằng cách giảm thiểu việc phân chia trang và thúc đẩy ghi tuần tự.20 |
| **Trường hợp sử dụng** | Tốt nhất cho các định danh mà tính không thể đoán trước là mục tiêu chính và hiệu năng ghi không phải là mối quan tâm chính. | Lý tưởng cho các khóa chính cơ sở dữ liệu trong các hệ thống ghi nhiều, cung cấp tính duy nhất toàn cầu của UUID với hiệu năng lập chỉ mục của các ID tuần tự. |

Hàm uuidv7() gốc của PostgreSQL 18 làm cho chiến lược tiên tiến này trở nên dễ dàng triển khai, mang lại lợi thế hiệu năng đáng kể so với UUID v4 cho các khối lượng công việc ghi nhiều điển hình của một máy chủ phân quyền.8
Đối với lớp ứng dụng Go, thư viện github.com/gofrs/uuid được khuyến nghị. Đây là một thư viện trưởng thành và được sử dụng rộng rãi, cung cấp hỗ trợ tạo UUID v7. Quan trọng nhất, kiểu UUID của nó triển khai các interface tiêu chuẩn database/sql.Scanner và driver.Valuer, cho phép tích hợp liền mạch với các trình điều khiển cơ sở dữ liệu như pgx hoặc lib/pq mà không cần mã chuyển đổi thủ công.21

## **Triển khai Hợp đồng Lưu trữ Fosite cho PostgreSQL**

Với lược đồ cơ sở dữ liệu đã được xác định, bước quan trọng tiếp theo là tạo một triển khai bằng Go đáp ứng các interface lưu trữ theo yêu cầu của thư viện Fosite. Triển khai này hoạt động như một cầu nối, chuyển đổi các yêu cầu của Fosite về việc lưu trữ dữ liệu thành các hoạt động SQL đối với cơ sở dữ liệu PostgreSQL. Phần này trình bày chi tiết việc tạo ra một SQLStore đáp ứng hợp đồng này, tập trung vào một cách tiếp cận mạnh mẽ và dễ bảo trì.

### **Hiểu về các Interface Lưu trữ của Fosite**

Fosite đạt được thiết kế không phụ thuộc vào lưu trữ thông qua một bộ interface Go toàn diện. Bất kỳ backend lưu trữ tùy chỉnh nào cũng phải cung cấp các triển khai cụ thể cho các interface này. SQLStore của chúng ta sẽ là một struct duy nhất triển khai tất cả các interface cần thiết, cung cấp một lớp truy cập dữ liệu thống nhất cho Fosite.24
Các interface chính phải được triển khai cho một máy chủ OAuth 2.0 và OpenID Connect đầy đủ tính năng là:

* fosite.ClientManager: Chịu trách nhiệm lấy thông tin client. Phương thức chính của nó là GetClient.
* oauth2.CoreStorage: Đây là một interface tổng hợp, nhúng ba interface quan trọng khác để quản lý vòng đời của mã và token:
    * oauth2.AuthorizeCodeStorage: Quản lý mã ủy quyền.
    * oauth2.AccessTokenStorage: Quản lý access token.
    * oauth2.RefreshTokenStorage: Quản lý refresh token.
* oauth2.PKCERequesterStorage: Xử lý việc lưu trữ và truy xuất các phiên yêu cầu PKCE.
* oidc.OpenIDConnectRequestStorage: Quản lý dữ liệu phiên dành riêng cho các luồng OpenID Connect.
* oauth2.TokenRevocationStorage: Cung cấp các phương thức để thu hồi access token và refresh token.

### **Thiết lập Struct SQLStore và các Phụ thuộc**

Struct SQLStore sẽ đóng vai trò là receiver cho tất cả các phương thức lưu trữ của chúng ta. Nó sẽ giữ một nhóm kết nối cơ sở dữ liệu và một hasher để so sánh client secret. Thư viện sqlx được khuyến nghị thay vì gói database/sql tiêu chuẩn vì nó cung cấp các tiện ích thuận tiện để quét kết quả truy vấn trực tiếp vào các struct Go và xử lý các tham số truy vấn được đặt tên, giúp giảm mã soạn sẵn.

```go
package storage

import (
"context"
"encoding/json"
"time"

    "github.com/gofrs/uuid/v5"
    "github.com/jmoiron/sqlx"
    "github.com/ory/fosite"
    "github.com/ory/fosite/handler/oauth2"
    "github.com/ory/fosite/handler/openid"
)

// SQLStore triển khai các interface lưu trữ của Fosite bằng backend PostgreSQL.
type SQLStore struct {
db     *sqlx.DB
hasher fosite.Hasher
}

// NewSQLStore tạo một SQLStore mới.
func NewSQLStore(db *sqlx.DB, hasher fosite.Hasher) *SQLStore {
return &SQLStore{
db:     db,
hasher: hasher,
}
}
```

### **Triển khai fosite.ClientManager**

Interface ClientManager có một phương thức chính, GetClient, mà Fosite gọi để tải cấu hình của một client trong một luồng phân quyền.
Việc triển khai truy vấn bảng oauth2_clients theo ID được cung cấp. Sau đó, nó điền vào một struct fosite.DefaultClient với dữ liệu được truy xuất. Một phần quan trọng của việc triển khai này là xử lý client secret. Cơ sở dữ liệu lưu trữ một bản băm của secret, không phải là văn bản gốc. Struct fosite.DefaultClient yêu cầu bản băm, và logic nội bộ của Fosite sau đó sẽ sử dụng hasher được cung cấp để so sánh nó với secret được cung cấp trong một yêu cầu.

```go
// GetClient lấy một client theo ID từ cơ sở dữ liệu.
func (s *SQLStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
var client fosite.DefaultClient
var redirectURIs, grantTypes, responseTypes, scopes string
var tenantID uuid.NullUUID

    query := `SELECT client_name, secret_hash, redirect_uris, grant_types, response_types, scopes, is_public, tenant_id FROM oauth2_clients WHERE id=$1`

    clientID, err := uuid.FromString(id)
    if err!= nil {
        return nil, fosite.ErrNotFound.WithWrap(err).WithDebug("Định dạng UUID không hợp lệ cho ID client")
    }

    err = s.db.QueryRowxContext(ctx, query, clientID).Scan(
        &client.Name,
        &client.Secret, // Đây là secret_hash từ DB
        pq.Array(&redirectURIs),
        pq.Array(&grantTypes),
        pq.Array(&responseTypes),
        pq.Array(&scopes),
        &client.Public,
        &tenantID,
    )
    if err!= nil {
        if err == sql.ErrNoRows {
            return nil, fosite.ErrNotFound.WithWrap(err)
        }
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    client.ID = id
    client.RedirectURIs = redirectURIs
    client.GrantTypes = grantTypes
    client.ResponseTypes = responseTypes
    client.Scopes = scopes

    // Bạn có thể thêm tenant_id vào siêu dữ liệu của client nếu logic ứng dụng của bạn cần
    // client.Metadata = map[string]any{"tenant_id": tenantID}

    return &client, nil
}
```

### **Triển khai oauth2.CoreStorage (Phương pháp Session Chung)**

Phần này thể hiện sức mạnh của thiết kế bảng oauth2_sessions chung. Các phương thức cho AuthorizeCodeStorage, AccessTokenStorage, và RefreshTokenStorage đều tuân theo một mẫu tương tự: tuần tự hóa fosite.Requester thành JSON để lưu trữ và giải tuần tự hóa nó khi truy xuất.

#### **Ví dụ: CreateAuthorizeCodeSession và GetAuthorizeCodeSession**

Hai phương thức này xử lý vòng đời của một mã ủy quyền. Mã được tạo, lưu trữ trong thời gian ngắn, và sau đó được truy xuất một lần để trao đổi.

```go
// Struct trợ giúp để tuần tự hóa
type sessionData struct {
Requester fosite.Requester `json:"requester"`
Session   json.RawMessage  `json:"session"`
}

// CreateAuthorizeCodeSession lưu trữ phiên mã ủy quyền trong cơ sở dữ liệu.
func (s *SQLStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
return s.createSession(ctx, signature, requester, "authorize_code")
}

// GetAuthorizeCodeSession truy xuất một phiên mã ủy quyền từ cơ sở dữ liệu.
func (s *SQLStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
return s.getSession(ctx, signature, session, "authorize_code")
}

// InvalidateAuthorizeCodeSession đánh dấu một phiên mã ủy quyền là không hoạt động.
func (s *SQLStore) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) error {
query := `UPDATE oauth2_sessions SET active = false WHERE signature = $1 AND session_type = 'authorize_code'`
_, err := s.db.ExecContext(ctx, query, signature)
if err!= nil {
return fosite.ErrServerError.WithWrap(err)
}
return nil
}
```

Các phương thức trợ giúp chung createSession và getSession đóng gói logic cốt lõi của việc tuần tự hóa và tương tác cơ sở dữ liệu. Mẫu này có khả năng tái sử dụng cao cho tất cả các tạo tác tạm thời.

```go
// createSession là một hàm trợ giúp chung để lưu trữ bất kỳ loại phiên nào.
func (s *SQLStore) createSession(ctx context.Context, signature string, requester fosite.Requester, sessionType string) error {
subjectID := uuid.NullUUID{}
if requester.GetSession()!= nil && requester.GetSession().GetSubject()!= "" {
subUUID, err := uuid.FromString(requester.GetSession().GetSubject())
if err == nil {
subjectID = uuid.NullUUID{UUID: subUUID, Valid: true}
}
}

    sessionBytes, err := json.Marshal(requester.GetSession())
    if err!= nil {
        return fosite.ErrServerError.WithWrap(err)
    }

    data := sessionData{
        Requester: requester,
        Session:   sessionBytes,
    }

    jsonData, err := json.Marshal(data)
    if err!= nil {
        return fosite.ErrServerError.WithWrap(err)
    }

    clientID, _ := uuid.FromString(requester.GetClient().GetID())

    query := `INSERT INTO oauth2_sessions (signature, request_id, session_type, session_data, expires_at, client_id, subject_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`
    _, err = s.db.ExecContext(ctx, query, signature, requester.GetID(), sessionType, jsonData, requester.GetRequestedAt().Add(requester.GetSession().GetExpiresAt(fosite.AuthorizeCode)), clientID, subjectID)

    if err!= nil {
        return fosite.ErrServerError.WithWrap(err)
    }
    return nil
}

// getSession là một hàm trợ giúp chung để truy xuất bất kỳ loại phiên nào.
func (s *SQLStore) getSession(ctx context.Context, signature string, session fosite.Session, sessionType string) (fosite.Requester, error) {
var rawDatabyte
var active bool

    query := `SELECT session_data, active FROM oauth2_sessions WHERE signature = $1 AND session_type = $2`
    err := s.db.QueryRowContext(ctx, query, signature, sessionType).Scan(&rawData, &active)
    if err!= nil {
        if err == sql.ErrNoRows {
            return nil, fosite.ErrNotFound.WithWrap(err)
        }
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    if\!active {
        // Fosite mong đợi requester được trả về ngay cả khi mã đã bị vô hiệu hóa.
        // Điều này cho phép nó phát hiện các cuộc tấn công phát lại.
        var data sessionData
        if err := json.Unmarshal(rawData, &data); err!= nil {
            return nil, fosite.ErrServerError.WithWrap(err)
        }
        return data.Requester, fosite.ErrInvalidatedAuthorizeCode
    }

    var data sessionData
    data.Requester = &fosite.Request{Session: session}
    if err := json.Unmarshal(rawData, &data); err!= nil {
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    return data.Requester, nil
}
```

Mẫu sử dụng các hàm trợ giúp chung để tuần tự hóa và truy cập cơ sở dữ liệu này có thể được mở rộng để triển khai AccessTokenStorage, RefreshTokenStorage, PKCERequesterStorage, và OpenIDConnectRequestStorage với mã bổ sung tối thiểu, thể hiện hiệu quả của thiết kế lược đồ đã chọn. Interface fosite.Requester là cấu trúc dữ liệu trung tâm nắm bắt toàn bộ trạng thái của một yêu cầu. Bằng cách coi nó như một đối tượng có thể tuần tự hóa, lớp lưu trữ trở thành một cơ chế lưu trữ đơn giản, tách biệt rõ ràng vai trò của cơ sở dữ liệu khỏi logic nghiệp vụ phức tạp do Fosite quản lý.

## **Điều phối các Luồng OAuth 2.0 với Gin Web Framework**

Với một lớp lưu trữ mạnh mẽ đã sẵn sàng, bước tiếp theo là phơi bày chức năng của máy chủ phân quyền thông qua các endpoint HTTP. Phần này trình bày chi tiết cách sử dụng framework web Gin để xử lý các yêu cầu đến và ủy quyền logic OAuth 2.0 phức tạp cho Fosite provider đã được cấu hình. Chìa khóa là cấu trúc các handler của Gin như những lớp bao bọc mỏng, điều phối sự tương tác giữa ngữ cảnh HTTP và công cụ xử lý của Fosite.

### **Khởi tạo và Cấu hình Fosite Provider**

Bước đầu tiên là khởi tạo chính Fosite provider. Điều này thường được thực hiện một lần khi ứng dụng khởi động, ví dụ, trong hàm main. Gói compose trong Fosite cung cấp các hàm trợ giúp tiện lợi để xây dựng một provider với các loại grant và tính năng khác nhau được bật.26
ComposeAllEnabled là một cách nhanh chóng để bật tất cả các handler OAuth 2.0 và OpenID Connect tiêu chuẩn. Hàm này kết nối các chiến lược và handler cần thiết, được cấu hình với triển khai lưu trữ tùy chỉnh và các tham số bảo mật của bạn.

```go
package main

import (
"crypto/rand"
"crypto/rsa"
"log"
"time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    "github.com/ory/fosite"
    "github.com/ory/fosite/compose"
    //... import gói lưu trữ tùy chỉnh của bạn
)

var (
// Secret này được sử dụng để ký và xác thực token.
// Trong môi trường production, hãy tải nó từ một nguồn cấu hình an toàn.
// Nó phải dài 32 byte cho HMAC-SHA256.
systemSecret =byte("some-super-secret-key-that-is-32-bytes")
)

func main() {
// 1. Kết nối đến cơ sở dữ liệu PostgreSQL của bạn
db, err := sqlx.Connect("pgx", "postgres://user:pass@host:port/db?sslmode=disable")
if err!= nil {
log.Fatalf("Không thể kết nối đến cơ sở dữ liệu: %v", err)
}

    // 2. Tạo một instance của SQLStore tùy chỉnh của bạn
    // Hasher BCrypt của Fosite là một lựa chọn mặc định tốt cho client secret.
    store := storage.NewSQLStore(db, &fosite.BCrypt{WorkFactor: 12})

    // 3. Cấu hình Fosite
    config := &fosite.Config{
        AccessTokenLifespan:   time.Hour * 1,
        AuthorizeCodeLifespan: time.Minute * 10,
        //... các cấu hình khác như tuổi thọ của refresh token
    }

    // Khóa này được sử dụng để ký ID Token của OpenID Connect.
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err!= nil {
        log.Fatalf("Không thể tạo khóa RSA: %v", err)
    }

    // 4. Khởi tạo Fosite provider
    oauth2Provider := compose.ComposeAllEnabled(config, store, systemSecret, privateKey)

    // 5. Thiết lập router và handler của Gin
    router := gin.Default()

    //... đăng ký các handler (xem phần tiếp theo)

    router.Run(":3000")
}
```

### **Cấu trúc các Handler của Gin cho các Endpoint OAuth 2.0**

Các handler của Gin đóng vai trò là giao diện giữa thế giới HTTP và Fosite. Trách nhiệm chính của chúng là phân tích http.Request đến, chuyển nó đến phương thức Fosite provider thích hợp, và sau đó sử dụng các hàm trợ giúp của Fosite để ghi phản hồi HTTP chính xác, cho dù đó là một payload JSON, một chuyển hướng, hay một lỗi.
Một nhóm router cho /oauth2 giúp tổ chức các endpoint một cách sạch sẽ.1

```go
// Trong hàm main của bạn hoặc một hàm thiết lập router riêng

// Tạo một struct auth handler để giữ provider
authHandler := &handler.AuthHandler{
Provider: oauth2Provider,
Store:    store, // Truyền store của bạn cho logic tùy chỉnh như xác thực người dùng
}

oauth2Group := router.Group("/oauth2")
{
oauth2Group.GET("/auth", authHandler.Authorize)
oauth2Group.POST("/auth", authHandler.Authorize) // Xử lý việc gửi biểu mẫu
oauth2Group.POST("/token", authHandler.Token)
oauth2Group.POST("/introspect", authHandler.Introspect)
oauth2Group.POST("/revoke", authHandler.Revoke)
}
```

#### **Handler authorize**

Đây là handler phức tạp nhất vì nó liên quan đến sự tương tác trực tiếp của người dùng: xác thực và chấp thuận. Các endpoint khác hoàn toàn là lập trình (kênh sau). Fosite không xử lý giao diện đăng nhập người dùng hoặc chấp thuận; đó là trách nhiệm của ứng dụng. Handler này phải kết nối việc quản lý phiên của ứng dụng với luồng phân quyền của Fosite.

```go
// Trong gói handler của bạn
type AuthHandler struct {
Provider fosite.OAuth2Provider
Store    *storage.SQLStore // Hoặc một dịch vụ có thể xác thực người dùng
}

func (h *AuthHandler) Authorize(c *gin.Context) {
ctx := c.Request.Context()

    // 1. Phân tích yêu cầu phân quyền
    ar, err := h.Provider.NewAuthorizeRequest(ctx, c.Request)
    if err!= nil {
        h.Provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
        return
    }

    // 2. Kiểm tra xem người dùng đã được xác thực chưa (sử dụng cookie phiên của ứng dụng bạn)
    userID, err := GetUserIDFromSession(c.Request) // Triển khai hàm này
    if err!= nil {
        // Chưa xác thực, chuyển hướng đến trang đăng nhập của bạn.
        // Truyền các tham số yêu cầu phân quyền ban đầu để bạn có thể tiếp tục sau khi đăng nhập.
        c.Redirect(http.StatusFound, "/login?return_to="+c.Request.URL.String())
        return
    }

    // 3. Xử lý sự chấp thuận của người dùng.
    // Nếu người dùng chưa chấp thuận cho client và các scope này:
    // - Hiển thị một trang chấp thuận hiển thị ar.GetClient().GetName() và ar.GetRequestedScopes().
    // - Biểu mẫu trang chấp thuận nên POST trở lại endpoint này.
    // Đối với ví dụ này, chúng ta sẽ giả định sự chấp thuận được cấp một cách ngầm định.

    // 4. Tạo một đối tượng phiên Fosite.
    // Chủ thể của phiên PHẢI là ID duy nhất của người dùng.
    session := &fosite.DefaultSession{
        Subject:  userID.String(),
        Username: "user-from-db", // Tùy chọn: lấy tên người dùng
    }

    // 5. Cấp các scope được yêu cầu.
    for _, scope := range ar.GetRequestedScopes() {
        ar.GrantScope(scope)
    }

    // 6. Để Fosite tạo phản hồi.
    response, err := h.Provider.NewAuthorizeResponse(ctx, ar, session)
    if err!= nil {
        h.Provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
        return
    }

    // 7. Ghi phản hồi cho user-agent.
    // Điều này thường sẽ là một Chuyển hướng 302 đến redirect_uri của client với một tham số `code`.
    h.Provider.WriteAuthorizeResponse(ctx, c.Writer, ar, response)
}
```

#### **Handler /token**

Handler này đơn giản hơn nhiều vì nó là một tương tác từ máy chủ đến máy chủ. Client POST mã ủy quyền (hoặc các thông tin xác thực khác dành riêng cho grant), và máy chủ phản hồi bằng các token.

```go
func (h *AuthHandler) Token(c *gin.Context) {
ctx := c.Request.Context()

    // 1. Tạo một đối tượng yêu cầu truy cập. Fosite sẽ xử lý việc phân tích phần thân của biểu mẫu.
    accessRequest, err := h.Provider.NewAccessRequest(ctx, c.Request, &fosite.DefaultSession{})
    if err!= nil {
        h.Provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
        return
    }

    // 2. Nếu đó là một grant refresh token, Fosite cần biết những scope nào đã được cấp ban đầu.
    if accessRequest.GetGrantTypes().ExactOne("refresh_token") {
        // Đối tượng phiên được truyền cho NewAccessRequest được Fosite điền vào thông qua việc tra cứu trong bộ lưu trữ.
        // Chúng ta chỉ cần cấp các scope từ yêu cầu ban đầu.
        for _, scope := range accessRequest.GetGrantedScopes() {
            accessRequest.GrantScope(scope)
        }
    }

    // 3. Để Fosite tạo phản hồi token.
    response, err := h.Provider.NewAccessResponse(ctx, accessRequest)
    if err!= nil {
        h.Provider.WriteAccessError(ctx, c.Writer, accessRequest, err)
        return
    }

    // 4. Ghi phản hồi JSON chứa các token.
    h.Provider.WriteAccessResponse(ctx, c.Writer, accessRequest, response)
}
```

#### **Các Handler introspect và revoke**

Các handler này tuân theo cùng một mẫu là ủy quyền trực tiếp cho Fosite provider. Chúng phải được bảo vệ để đảm bảo rằng chỉ các client đã được xác thực mới có thể kiểm tra hoặc thu hồi token. Việc bảo vệ này có thể được triển khai bằng cách sử dụng một middleware của Gin thực hiện Xác thực Cơ bản HTTP bằng thông tin đăng nhập của client.

```go
func (h *AuthHandler) Introspect(c *gin.Context) {
ctx := c.Request.Context()
ir, err := h.Provider.NewIntrospectionRequest(ctx, c.Request, &fosite.DefaultSession{})
if err!= nil {
h.Provider.WriteIntrospectionError(c.Writer, err)
return
}
h.Provider.WriteIntrospectionResponse(c.Writer, ir)
}

func (h *AuthHandler) Revoke(c *gin.Context) {
ctx := c.Request.Context()
err := h.Provider.NewRevocationRequest(ctx, c.Request)
if err!= nil {
h.Provider.WriteRevocationResponse(c.Writer, err)
return
}
h.Provider.WriteRevocationResponse(c.Writer, nil)
}
```

## **Tận dụng Redis để Tăng hiệu năng và Khả năng mở rộng**

Để xây dựng một máy chủ phân quyền hiệu năng cao, việc giảm tải các hoạt động tạm thời, có thông lượng cao khỏi cơ sở dữ liệu PostgreSQL chính là rất quan trọng. Redis, với vai trò là một kho dữ liệu trong bộ nhớ, là công cụ lý tưởng cho mục đích này. Phần này trình bày chi tiết việc triển khai một lớp lưu trữ dựa trên Redis cho dữ liệu tạm thời của Fosite và phác thảo một chiến lược mạnh mẽ để thu hồi token gần như ngay lập tức.

### **Triển khai Fosite Store sử dụng Redis**

Chiến lược lưu trữ lai bao gồm việc tạo ra một RedisStore chuyên dụng, triển khai các interface của Fosite cho dữ liệu tạm thời. Kho lưu trữ này sẽ xử lý các tạo tác như mã ủy quyền và các phiên PKCE, vốn có vòng đời ngắn, được xác định rõ ràng.
Đầu tiên, định nghĩa struct RedisStore, sẽ giữ một kết nối client từ một thư viện như go-redis/v9.29

```go
package storage

import (
    "context"
    "encoding/json"
    "errors"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/ory/fosite"
)

// RedisStore triển khai các interface lưu trữ của Fosite cho dữ liệu tạm thời.
type RedisStore struct {
    client *redis.Client
}

// NewRedisStore tạo một RedisStore mới.
func NewRedisStore(client *redis.Client) *RedisStore {
    return &RedisStore{client: client}
}
```

Tiếp theo, triển khai các interface Fosite có liên quan. Mẫu này nhất quán: tuần tự hóa đối tượng fosite.Requester thành JSON và sử dụng lệnh SET của Redis với thời gian hết hạn (EX) để lưu trữ nó. Khóa nên có tiền tố để tránh xung đột và cung cấp không gian tên.

#### **Ví dụ: Triển khai AuthorizeCodeStorage trong Redis**

```go
const (
    authCodeKeyPrefix = "fosite:auth_code:"
)

func (s *RedisStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
key := authCodeKeyPrefix + signature

    lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
    if lifespan <= 0 {
        return fosite.ErrServerError.WithDebug("Tuổi thọ mã ủy quyền bằng không hoặc âm.")
    }

    data, err := json.Marshal(requester)
    if err!= nil {
        return fosite.ErrServerError.WithWrap(err)
    }

    return s.client.Set(ctx, key, data, lifespan).Err()
}

func (s *RedisStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
key := authCodeKeyPrefix + signature

    data, err := s.client.Get(ctx, key).Bytes()
    if err!= nil {
        if errors.Is(err, redis.Nil) {
            return nil, fosite.ErrNotFound.WithWrap(err)
        }
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    // Requester cần một đối tượng session để giải tuần tự hóa vào.
    requester := &fosite.Request{Session: session}
    if err := json.Unmarshal(data, requester); err!= nil {
        return nil, fosite.ErrServerError.WithWrap(err)
    }

    return requester, nil
}

func (s *RedisStore) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) error {
    key := authCodeKeyPrefix + signature
    // Xóa khóa thực chất là vô hiệu hóa nó.
    return s.client.Del(ctx, key).Err()
}
```

Mẫu tương tự này có thể được áp dụng để triển khai PKCERequesterStorage và các interface lưu trữ tạm thời khác do Fosite cung cấp.31

### **Chiến lược Thu hồi và Đưa Token vào Danh sách đen**

Mặc dù access token được thiết kế để có tuổi thọ ngắn, có những trường hợp một token phải được thu hồi *ngay lập tức*, trước khi hết hạn tự nhiên. Ví dụ bao gồm người dùng đăng xuất, thay đổi mật khẩu, hoặc quản trị viên vô hiệu hóa tài khoản. Dựa vào việc kiểm tra cơ sở dữ liệu cho mỗi lệnh gọi API để xác thực một token là không hiệu quả. Một danh sách đen dựa trên Redis là giải pháp tiêu chuẩn ngành cho vấn đề này.12
Chiến lược rất đơn giản:

1. Khi một token được cấp, nó nên chứa một định danh duy nhất. Đối với JWT, đây là claim jti (JWT ID). Đối với các token không tường minh (opaque token), chữ ký của token có thể đóng vai trò là ID duy nhất của nó.
2. Khi một token bị thu hồi (ví dụ, thông qua endpoint /oauth2/revoke), định danh duy nhất này được thêm vào một tập hợp Redis hoặc được lưu trữ dưới dạng một khóa.
3. Khóa trong Redis nên được đặt thời gian hết hạn bằng với tuổi thọ ban đầu của token. Điều này đảm bảo danh sách đen không phát triển vô hạn với các token đã hết hạn.
4. Trong quá trình xác thực token (ví dụ, trong middleware của một máy chủ tài nguyên hoặc endpoint /oauth2/introspect), hệ thống phải thực hiện một kiểm tra nhanh chóng đối với Redis để xem định danh của token có nằm trong danh sách đen hay không. Việc kiểm tra này cực kỳ nhanh và xảy ra trước bất kỳ xác thực mật mã nào.32

#### **Triển khai Logic Thu hồi**

Interface TokenRevocationStorage trong Fosite có thể được triển khai để sử dụng danh sách đen Redis này.

```go
const (
    revokedTokenKeyPrefix = "fosite:revoked:"
)

// RevokeAccessToken thêm ID yêu cầu của một access token vào danh sách đen Redis.
func (s *RedisStore) RevokeAccessToken(ctx context.Context, requestID string) error {
    // Điều này đòi hỏi một tra cứu riêng để lấy thời gian hết hạn của token,
    // hoặc một quy ước trong đó chính requestID chứa đủ thông tin.
    // Một triển khai mạnh mẽ hơn sẽ lưu trữ thời gian hết hạn cùng với phiên.
    // Để đơn giản, chúng ta sẽ sử dụng một giới hạn trên cố định, an toàn.
    return s.client.Set(ctx, revokedTokenKeyPrefix+requestID, "revoked", time.Hour*24).Err()
}

// RevokeRefreshToken thêm ID yêu cầu của một refresh token vào danh sách đen Redis.
func (s *RedisStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
    return s.client.Set(ctx, revokedTokenKeyPrefix+requestID, "revoked", time.Hour*24*30).Err()
}
```

Để làm cho điều này thực sự hiệu quả, logic kiểm tra (introspection) phải được cập nhật để kiểm tra danh sách đen này.

### **HybridStore: Một Triển khai Tổng hợp**

Để tích hợp cả SQLStore và RedisStore một cách liền mạch với Fosite, vốn mong đợi một đối tượng lưu trữ duy nhất, một struct HybridStore tổng hợp có thể được sử dụng. Mẫu này tận dụng tính năng nhúng struct của Go để kết hợp các triển khai.

```go
package storage

// HybridStore kết hợp các kho lưu trữ SQL và Redis.
// Nó ủy quyền các lệnh gọi đến kho lưu trữ được nhúng thích hợp.
type HybridStore struct {
    *SQLStore
    *RedisStore
}

func NewHybridStore(db *sqlx.DB, client *redis.Client, hasher fosite.Hasher) *HybridStore {
    return &HybridStore{
        SQLStore:   NewSQLStore(db, hasher),
        RedisStore: NewRedisStore(client),
    }
}
```

Trong cấu trúc này, RedisStore sẽ triển khai các interface cho dữ liệu tạm thời (AuthorizeCodeStorage, PKCERequesterStorage), trong khi SQLStore sẽ triển khai các interface cho dữ liệu bền vững (ClientManager, RefreshTokenStorage nếu refresh token được lưu trữ lâu dài). Khi một phương thức như CreateAuthorizeCodeSession được gọi trên HybridStore, cơ chế phân giải phương thức của Go sẽ tìm thấy nó trên *RedisStore được nhúng. Khi GetClient được gọi, nó sẽ được tìm thấy trên *SQLStore được nhúng. Điều này tạo ra một backend lưu trữ sạch sẽ, mạnh mẽ và được tối ưu hóa cao, trình bày một giao diện duy nhất, thống nhất cho Fosite, hiện thực hóa hoàn hảo chiến lược lưu trữ lai.

## **Luồng Hệ thống End-to-End: Authorization Code Grant với PKCE**

Để tổng hợp các thành phần kiến trúc và chi tiết triển khai đã thảo luận, phần này cung cấp một hướng dẫn toàn diện về luồng OAuth 2.0 phổ biến và an toàn nhất: Authorization Code Grant với Proof Key for Code Exchange (PKCE). Luồng này là tiêu chuẩn được khuyến nghị cho cả ứng dụng web và di động.35 Câu chuyện theo sau một yêu cầu từ khi bắt đầu đến khi cấp token và xác thực sau đó, làm nổi bật vai trò của từng thành phần—Gin, Fosite, PostgreSQL và Redis—ở mỗi bước.

### **Sơ đồ Tuần tự**

Sơ đồ tuần tự sau đây minh họa các tương tác chi tiết giữa User-Agent (trình duyệt), Ứng dụng Client, Máy chủ Phân quyền (ứng dụng Gin/Fosite của chúng ta), và các kho dữ liệu (Redis và PostgreSQL).

Đoạn mã

```diagram
sequenceDiagram
participant UA as User-Agent
participant Client as Ứng dụng Client
participant Server as Máy chủ Phân quyền (Gin/Fosite)
participant Redis
participant PG as PostgreSQL

    autonumber

    Client->>Client: Tạo code_verifier & code_challenge
    Client->>UA: Chuyển hướng đến /oauth2/auth
    UA->>Server: GET /oauth2/auth?client_id=...&code_challenge=...
    Server->>PG: GetClient(client_id)
    Server->>UA: Hiển thị Giao diện Đăng nhập/Chấp thuận
    UA->>Server: POST thông tin xác thực & chấp thuận
    Server->>PG: Xác thực người dùng
    Server->>Server: Tạo Authorization Code
    Server->>Redis: CreateAuthorizeCodeSession(code, requester_with_challenge)
    Server->>UA: Chuyển hướng đến client_app?code=...
    UA->>Client: Gửi authorization code
    Client->>Server: POST /oauth2/token (code, client_secret, code_verifier)
    Server->>Redis: GetAuthorizeCodeSession(code)
    Server->>Redis: InvalidateAuthorizeCodeSession(code)
    Server->>PG: GetClient(client_id) để xác minh secret
    Server->>Server: Xác thực code_verifier với challenge đã lưu
    Server->>Server: Cấp Access & Refresh Token
    Server->>Redis: CreateAccessTokenSession(access_token_sig)
    Server->>Client: Phản hồi với {access_token, refresh_token}

    Note over Client, PG: --- Sau đó, trong khi gọi API ---

    Client->>Server: Gọi API của Resource Server với Access Token
    Note right of Client: (Resource Server không được hiển thị)
    Server->>Server: Kiểm tra Token (ví dụ, qua middleware)
    Server->>Redis: Kiểm tra chữ ký token trong danh sách đen thu hồi
    Server->>Redis: GetAccessTokenSession(access_token_sig)
    alt Token không có trong Redis
        Server->>PG: Xác thực dự phòng (nếu có)
    end
    Server->>Client: Xử lý yêu cầu API
```


### **Hướng dẫn Tường thuật**

Phần tường thuật từng bước này tương ứng với sơ đồ tuần tự, cung cấp ngữ cảnh cho mỗi hành động.

1. **Tạo PKCE (Phía Client)**: Quá trình bắt đầu trong ứng dụng client. Trước khi bắt đầu luồng, nó tạo ra một chuỗi ngẫu nhiên về mặt mật mã, code_verifier. Sau đó, nó tạo ra một code_challenge bằng cách băm verifier (thường sử dụng SHA-256) và mã hóa kết quả bằng Base64-URL.
2. **Yêu cầu Phân quyền (Kênh trước)**: Client chuyển hướng trình duyệt của người dùng đến endpoint /oauth2/auth của máy chủ phân quyền. Yêu cầu bao gồm các tham số như client_id, response_type=code, scope, redirect_uri, và code_challenge và code_challenge_method đã tạo.
3. **Phân tích Yêu cầu và Xác thực Client**: Handler Gin cho /oauth2/auth nhận yêu cầu. Nó chuyển http.Request đến fosite.Provider.NewAuthorizeRequest. Fosite phân tích tất cả các tham số và gọi phương thức GetClient trên triển khai lưu trữ của chúng ta. SQLStore truy vấn bảng oauth2_clients trong PostgreSQL để lấy chi tiết của client và xác thực rằng redirect_uri đã được đăng ký.
4. **Xác thực và Chấp thuận của Người dùng**: Handler Gin xác định rằng người dùng chưa được xác thực (ví dụ: không có cookie phiên ứng dụng hợp lệ). Nó chuyển hướng người dùng đến một trang đăng nhập. Sau khi đăng nhập thành công (bao gồm việc xác thực thông tin đăng nhập với bảng users trong PostgreSQL), người dùng được trình bày một màn hình chấp thuận, hiển thị tên của client và các quyền được yêu cầu (scopes).
5. **Tạo Mã ủy quyền**: Khi người dùng chấp thuận, handler Gin tạo một fosite.Session chứa ID của người dùng đã xác thực. Sau đó, nó gọi fosite.Provider.NewAuthorizeResponse. Bên trong, Fosite tạo ra một mã ủy quyền duy nhất, tồn tại trong thời gian ngắn.
6. **Lưu trữ Phiên trong Redis**: Fosite gọi phương thức CreateAuthorizeCodeSession trên triển khai lưu trữ của chúng ta. HybridStore định tuyến điều này đến RedisStore. RedisStore tuần tự hóa toàn bộ trạng thái yêu cầu (bao gồm cả code_challenge) thành JSON và lưu trữ nó trong Redis với một khóa được lấy từ chữ ký của mã ủy quyền, đặt thời gian hết hạn là vài phút.
7. **Chuyển hướng đến Client (Kênh trước)**: WriteAuthorizeResponse của Fosite gửi một phản hồi chuyển hướng 302 Found đến trình duyệt, hướng nó trở lại redirect_uri của client với mã ủy quyền code được nối vào như một tham số truy vấn.
8. **Yêu cầu Trao đổi Token (Kênh sau)**: Ứng dụng client nhận code từ URL. Sau đó, nó thực hiện một yêu cầu POST trực tiếp, từ máy chủ đến máy chủ đến endpoint /oauth2/token. Phần thân yêu cầu chứa grant_type=authorization_code, code, redirect_uri, client_id, client_secret của nó, và code_verifier văn bản gốc ban đầu.
9. **Xác thực và Tiêu thụ Mã**: Handler /token của Gin gọi fosite.Provider.NewAccessRequest. Fosite gọi GetAuthorizeCodeSession trên lớp lưu trữ. RedisStore truy xuất phiên đã được tuần tự hóa từ Redis bằng chữ ký của mã. Nếu tìm thấy, Fosite sau đó gọi InvalidateAuthorizeCodeSession, khiến RedisStore xóa khóa khỏi Redis, đảm bảo mã chỉ được sử dụng một lần.
10. **Xác thực Client và PKCE**: Fosite xác minh client_secret bằng cách so sánh bản băm từ bảng oauth2_clients (được truy xuất từ PostgreSQL) với secret được cung cấp. Sau đó, nó băm code_verifier từ yêu cầu và so sánh nó với code_challenge được truy xuất từ phiên Redis, ngăn chặn các cuộc tấn công chặn mã ủy quyền.
11. **Cấp Token**: Với tất cả các xác thực đã qua, Fosite tạo ra một access token mới và một refresh token. Nó có thể gọi CreateAccessTokenSession và CreateRefreshTokenSession trên lớp lưu trữ. HybridStore có thể được cấu hình để định tuyến các lệnh gọi này đến Redis để kiểm tra nhanh hoặc đến PostgreSQL để lưu trữ refresh token bền vững.
12. **Phản hồi Token**: WriteAccessResponse của Fosite gửi một phản hồi 200 OK cho client với một phần thân JSON chứa access_token, refresh_token, expires_in, và token_type.
13. **Kiểm tra Token (Trong khi gọi API)**: Sau đó, client sử dụng access_token để thực hiện một lệnh gọi đến một máy chủ tài nguyên được bảo vệ. Middleware của máy chủ tài nguyên gọi endpoint /oauth2/introspect trên máy chủ phân quyền của chúng ta. Handler kiểm tra trước tiên kiểm tra danh sách đen Redis cho chữ ký/JTI của token. Nếu không tìm thấy, nó kiểm tra phiên hoạt động trong Redis. Việc kiểm tra hai bước này trên Redis cung cấp một đường dẫn rất nhanh để xác thực token. Chỉ khi token không được tìm thấy trong Redis, nó mới có thể quay lại kiểm tra cơ sở dữ liệu chậm hơn, tùy thuộc vào chiến lược lưu trữ cho access token.

Luồng chi tiết này cho thấy cách các thành phần hoạt động phối hợp để cung cấp một quy trình phân quyền an toàn và hiệu năng, sử dụng PostgreSQL một cách thông minh cho trạng thái bền vững và Redis cho các hoạt động tạm thời tốc độ cao.

## **Tăng cường Bảo mật và Các Lưu ý cho Môi trường Production**

Triển khai một máy chủ phân quyền vào môi trường production đòi hỏi sự tập trung nghiêm ngặt vào bảo mật, khả năng mở rộng và khả năng bảo trì. Thiết kế kiến trúc cung cấp một nền tảng vững chắc, nhưng một số thực tiễn chính phải được triển khai để tăng cường hệ thống chống lại các mối đe dọa và đảm bảo nó hoạt động đáng tin cậy dưới tải.

### **Quản lý Client Secret**

client_secret là một thông tin xác thực rất nhạy cảm, xác thực ứng dụng client với máy chủ phân quyền. Việc quản lý sai lầm nó có thể dẫn đến các vi phạm bảo mật nghiêm trọng.

* **Băm, không Mã hóa**: Client secret **không bao giờ** được lưu trữ dưới dạng văn bản gốc hoặc ở định dạng có thể giải mã ngược. Phương pháp an toàn duy nhất là lưu trữ một bản băm mật mã một chiều của secret. Khi một client xác thực, máy chủ sẽ băm secret được cung cấp và so sánh nó với bản băm được lưu trữ.36
* **Lựa chọn Thuật toán**: Sử dụng một thuật toán băm mật khẩu mạnh, được thiết kế để chậm và chống lại các cuộc tấn công brute-force. **BCrypt** là một lựa chọn đã được thiết lập và an toàn. Fosite cung cấp một triển khai fosite.BCrypt tích hợp sẵn tuân thủ interface fosite.Hasher của nó, giúp việc tích hợp trở nên đơn giản.5 Hệ số công việc (work factor) cho BCrypt nên được cấu hình càng cao càng tốt mà hiệu năng cho phép (ví dụ: 10-14).
* **Lưu trữ và Phân phối An toàn**: Ngoài cơ sở dữ liệu, client secret phải được xử lý an toàn trong suốt vòng đời của chúng. Chúng không nên được mã hóa cứng trong mã nguồn ứng dụng, cam kết vào kiểm soát phiên bản, hoặc bị lộ trong nhật ký.37 Các secret trong môi trường production nên được quản lý bằng một hệ thống quản lý secret chuyên dụng (ví dụ: HashiCorp Vault, AWS Secrets Manager, Google Secret Manager) và được đưa vào môi trường ứng dụng khi chạy.39

### **Quản lý Vòng đời Token**

Quản lý đúng đắn tuổi thọ và việc thu hồi token là rất quan trọng để cân bằng giữa bảo mật và trải nghiệm người dùng.

* **Access Token có Tuổi thọ Ngắn**: Access token cấp quyền truy cập trực tiếp vào tài nguyên và do đó có nguy cơ cao nhất nếu bị xâm phạm. Tuổi thọ của chúng nên được giữ càng ngắn càng tốt cho trường hợp sử dụng của ứng dụng, thường là từ 15 đến 60 phút. Điều này giới hạn cửa sổ cơ hội cho kẻ tấn công sử dụng một token bị đánh cắp.41
* **Refresh Token có Tuổi thọ Dài hơn**: Refresh token cung cấp trải nghiệm người dùng tốt hơn bằng cách cho phép client lấy access token mới mà không buộc người dùng phải xác thực lại. Chúng có thể có tuổi thọ dài hơn nhiều, chẳng hạn như vài ngày hoặc vài tuần. Tuy nhiên, chúng phải được client lưu trữ an toàn (ví dụ: trong một cơ sở dữ liệu được mã hóa trên máy chủ, hoặc bộ nhớ an toàn trên thiết bị di động) và không bao giờ được phơi bày trong môi trường trình duyệt.43
* **Xoay vòng Refresh Token (Refresh Token Rotation)**: Đây là một cơ chế bảo mật quan trọng nên được bật. Khi một refresh token được sử dụng để lấy một access token mới, máy chủ phân quyền sẽ vô hiệu hóa refresh token đã sử dụng và cấp một cái mới. Điều này đảm bảo rằng mỗi refresh token chỉ được sử dụng một lần, giảm thiểu nguy cơ tấn công phát lại nếu một refresh token bị đánh cắp. Fosite hỗ trợ tính năng này thông qua cấu hình của nó.41
* **Thu hồi Ngay lập tức**: Như đã trình bày chi tiết trong phần Redis, một cơ chế thu hồi hiệu quả sử dụng danh sách đen là cần thiết để vô hiệu hóa token ngay lập tức khi có các sự kiện bảo mật như người dùng đăng xuất hoặc thay đổi mật khẩu.

### **Khả năng mở rộng và Đánh chỉ mục Cơ sở dữ liệu**

Khi số lượng người dùng, khách hàng và client tăng lên, hiệu năng cơ sở dữ liệu sẽ trở thành một yếu tố quan trọng. Mặc dù chiến lược lưu trữ lai giảm tải phần lớn lưu lượng đọc nhiều, việc đánh chỉ mục đúng cách trong PostgreSQL vẫn cần thiết cho các hoạt động thực sự truy cập cơ sở dữ liệu.

* **Khóa chính**: Việc sử dụng UUID v7 làm khóa chính cho tất cả các bảng cung cấp hiệu năng lập chỉ mục tuyệt vời cho các lần chèn và các truy vấn phạm vi dựa trên thời gian ngay từ đầu.
* **Khóa ngoại**: Tất cả các cột khóa ngoại nên được đánh chỉ mục để đảm bảo hiệu năng JOIN nhanh. Hầu hết các phiên bản PostgreSQL hiện đại đều tự động làm điều này, nhưng cần phải xác minh.
* **Chỉ mục Dành riêng cho Truy vấn**: Phân tích các mẫu truy vấn phổ biến và thêm các chỉ mục tương ứng.
    * users(email): Một chỉ mục duy nhất trên cột email là rất quan trọng để tra cứu nhanh trong quá trình đăng nhập.
    * user_tenant_memberships(tenant_id, user_id): Một chỉ mục tổng hợp sẽ tăng tốc độ kiểm tra tư cách thành viên của người dùng trong một khách hàng.
    * oauth2_sessions(request_id): Nếu refresh token được lưu trữ trong cơ sở dữ liệu, một chỉ mục trên request_id là rất quan trọng để tra cứu nhanh trong quá trình thu hồi.
    * oauth2_clients(tenant_id): Một chỉ mục trên tenant_id sẽ cần thiết để liệt kê tất cả các client thuộc về một khách hàng cụ thể.

### **Cấu trúc Dự án cho Ứng dụng Go Cấp Doanh nghiệp**

Một cấu trúc dự án được tổ chức tốt là tối quan trọng cho khả năng bảo trì lâu dài, đặc biệt là trong một ứng dụng cấp doanh nghiệp. "Bố cục Dự án Go Tiêu chuẩn" cung cấp một điểm khởi đầu đã được cộng đồng kiểm chứng, thúc đẩy sự tách biệt rõ ràng các mối quan tâm.7
Một cấu trúc được đề xuất cho máy chủ phân quyền này sẽ là:

```
/system
├── cmd/
│   └── server/
│       └── main.go
|
├── configs/
│   └── config.go
|
├── internal/
│   ├── app/
│   │   ├── handler/
│   │   │   └── v1/
│   │   │       ├── oauth2/            // MỚI: Chứa các handler cho endpoint OAuth2 (/authorize, /token,...)
│   │   │       │   ├── handler.go
│   │   │       │   └── dto.go
│   │   │       ├── user/
│   │   │       │   ├── handler.go
│   │   │       │   └── dto.go
│   │   │       └── product/
│   │   │           ├── handler.go
│   │   │           └── dto.go
│   │   └── middleware/
│   │       ├── auth.go              // Middleware hiện có
│   │       └── oauth_client.go      // MỚI: Middleware để xác thực client (ví dụ cho /introspect)
│   ├── domain/
│   │   ├── user.go
│   │   ├── product.go
│   │   ├── tenant.go                // MỚI: Entity và Interfaces cho Tenant
│   │   ├── role.go                  // MỚI: Entity và Interfaces cho Role
│   │   └── permission.go            // MỚI: Entity và Interfaces cho Permission
│   ├── oauth2/                      // MỚI: Gói chuyên dụng cho logic OAuth2 cốt lõi
│   │   ├── provider.go              // MỚI: Nơi khởi tạo và cấu hình Fosite provider
│   │   └── storage/                 // MỚI: Triển khai các interface lưu trữ của Fosite
│   │       ├── hybrid_store.go      // MỚI: Store tổng hợp (SQL + Redis)
│   │       ├── sql_store.go         // MỚI: Triển khai lưu trữ PostgreSQL cho Fosite
│   │       └── redis_store.go       // MỚI: Triển khai lưu trữ Redis cho Fosite
│   ├── pkg/
│   │   ├── service/
│   │   │   ├── user.go
│   │   │   ├── product.go
│   │   │   ├── tenant_service.go    // MỚI: Triển khai logic nghiệp vụ cho Tenant
│   │   │   └── rbac_service.go      // MỚI: Triển khai logic nghiệp vụ cho Role/Permission
│   │   └── repository/
│   │       ├── user_repo.go
│   │       ├── product_repo.go
│   │       ├── tenant_repo.go       // MỚI: Triển khai repo cho Tenant
│   │       └── rbac_repo.go         // MỚI: Triển khai repo cho Role/Permission
│   └── platform/
│       ├── database/
│       │   ├── postgres.go
│       │   └── redis.go             // MỚI: Logic kết nối/khởi tạo Redis client
│       └── logger/
│           └── zap.go
|
├── migrations/
│   ├──...
│   └── 000002_create_oauth_tables.up.sql   // MỚI: Migration cho các bảng OAuth2 và RBAC
|
├── pkg/
│   └── utils/
│       └── validation.go
|
├──.env
├── go.mod
├── go.sum
└── Dockerfile
```

Cấu trúc này tách biệt rõ ràng các mối quan tâm: cmd chứa tệp thực thi, internal chứa tất cả mã dành riêng cho ứng dụng không nhằm mục đích được các dự án khác nhập, và pkg có thể chứa bất kỳ thư viện nào thực sự có thể tái sử dụng.47 Trong internal, mã được tổ chức thêm theo chức năng: api cho lớp web, storage cho lớp lưu trữ, và auth cho thiết lập Fosite cốt lõi. Tính mô-đun này làm cho hệ thống dễ hiểu, kiểm thử và bảo trì hơn khi nó phát triển.

Giải thích Chi tiết về các Thành phần Mới

internal/oauth2/: Đây là một gói mới, chuyên dụng để chứa tất cả logic cốt lõi liên quan đến OAuth2. Việc tách nó ra khỏi app, domain hay platform giúp giữ cho hệ thống phân quyền được đóng gói và dễ quản lý.

provider.go: Tệp này sẽ chứa hàm khởi tạo và cấu hình Fosite provider (compose.ComposeAllEnabled). Nó sẽ nhận vào cấu hình và một storage (chính là HybridStore) để thiết lập Fosite.

storage/: Thư mục con này chứa các triển khai cụ thể cho các interface lưu trữ mà Fosite yêu cầu.

sql_store.go: Triển khai các interface cho dữ liệu bền vững (như ClientManager) bằng PostgreSQL.

redis_store.go: Triển khai các interface cho dữ liệu tạm thời (như AuthorizeCodeStorage, PKCERequesterStorage) bằng Redis.

hybrid_store.go: Một struct tổng hợp, nhúng SQLStore và RedisStore để cung cấp một đối tượng lưu trữ duy nhất cho Fosite, hiện thực hóa chiến lược lưu trữ lai.

internal/app/handler/v1/oauth2/: Theo đúng cấu trúc của bạn, đây là nơi để các HTTP handler cho các endpoint OAuth2.

handler.go: Sẽ chứa một struct OAuth2Handler với các phương thức như Authorize, Token, Introspect, Revoke. Handler này sẽ nhận fosite.OAuth2Provider đã được khởi tạo và gọi các phương thức tương ứng của nó.

dto.go: Chứa các Data Transfer Object nếu cần, ví dụ như struct để bind dữ liệu từ form của trang đăng nhập/chấp thuận.

internal/domain/: Mở rộng lớp domain của bạn với các khái niệm RBAC và multi-tenancy.

tenant.go, role.go, permission.go: Các tệp này sẽ định nghĩa các struct Tenant, Role, Permission (entity) và các interface tương ứng cho service và repository (ví dụ: TenantRepository, RoleService).

internal/pkg/: Triển khai các interface từ lớp domain.

repository/tenant_repo.go, rbac_repo.go: Chứa code truy vấn cơ sở dữ liệu PostgreSQL cho các entity mới.

service/tenant_service.go, rbac_service.go: Chứa logic nghiệp vụ liên quan đến quản lý tenant, vai trò và quyền.

internal/platform/database/redis.go: Tương tự như postgres.go, tệp này sẽ chịu trách nhiệm khởi tạo và quản lý kết nối đến Redis, trả về một *redis.Client.

migrations/: Bạn sẽ tạo một tệp migration mới (ví dụ: 000002_create_oauth_tables.up.sql) để định nghĩa tất cả các bảng đã được thiết kế trong tài liệu trước (tenants, users, roles, permissions, các bảng nối, oauth2_clients, oauth2_sessions).

Luồng Khởi tạo trong cmd/server/main.go

Điểm khởi đầu của bạn sẽ được cập nhật để "kết nối" tất cả các thành phần này lại với nhau (Dependency Injection).

```go
// cmd/server/main.go

func main() {
    // 1. Tải cấu hình (bao gồm cả chuỗi kết nối DB và Redis)
    cfg := configs.LoadConfig()

    // 2. Khởi tạo logger
    logger := logger.NewZapLogger(cfg)

    // 3. Khởi tạo kết nối cơ sở dữ liệu
    pgDB := database.NewPostgresConnection(cfg.PostgresDSN)
    redisClient := database.NewRedisConnection(cfg.RedisAddr) // MỚI

    // 4. Khởi tạo các Repository
    userRepo := repository.NewUserRepository(pgDB)
    //... các repo khác
    // MỚI: Khởi tạo các repo cho RBAC và Fosite storage
    fositeSQLStore := storage.NewSQLStore(pgDB, &fosite.BCrypt{}) // Hasher cho client secret
    fositeRedisStore := storage.NewRedisStore(redisClient)
    hybridStore := storage.NewHybridStore(fositeSQLStore, fositeRedisStore)

    // 5. Khởi tạo các Service
    userService := service.NewUserService(userRepo)
    //... các service khác

    // 6. Khởi tạo Fosite Provider (MỚI)
    oauth2Provider := oauth2.NewOAuth2Provider(cfg.OAuth2, hybridStore)

    // 7. Khởi tạo các Handler
    userHandler := v1.NewUserHandler(userService)
    // MỚI: Khởi tạo handler cho OAuth2
    oauth2Handler := v1_oauth2.NewOAuth2Handler(oauth2Provider, userService) // userService để xác thực người dùng

    // 8. Thiết lập Gin Router và đăng ký các route
    router := gin.Default()

    // Đăng ký các route hiện có
    //...

    // MỚI: Đăng ký các route cho OAuth2
    oauth2Group := router.Group("/oauth2")
    {
        oauth2Group.GET("/auth", oauth2Handler.Authorize)
        oauth2Group.POST("/auth", oauth2Handler.Authorize)
        oauth2Group.POST("/token", oauth2Handler.Token)
        // Thêm middleware xác thực client cho các endpoint cần thiết
        oauth2Group.POST("/introspect", middleware.ClientAuth(hybridStore), oauth2Handler.Introspect)
        oauth2Group.POST("/revoke", middleware.ClientAuth(hybridStore), oauth2Handler.Revoke)
    }

    // 9. Chạy server
    router.Run(":" + cfg.ServerPort)
}
```

#### **Nguồn trích dẫn**

1. Gin Web Framework, truy cập vào tháng 10 26, 2025, [https://gin-gonic.com/](https://gin-gonic.com/)
2. gin-gonic/gin: Gin is a high-performance HTTP web framework written in Go. It provides a Martini-like API but with significantly better performance—up to 40 times faster—thanks to httprouter. Gin is designed for building REST APIs, web applications, and microservices. - GitHub, truy cập vào tháng 10 26, 2025, [https://github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
3. fosite package - github.com/fengshch/fosite - Go Packages, truy cập vào tháng 10 26, 2025, [https://pkg.go.dev/github.com/fengshch/fosite](https://pkg.go.dev/github.com/fengshch/fosite)
4. ory/fosite: Extensible security first OAuth 2.0 and OpenID Connect SDK for Go. - GitHub, truy cập vào tháng 10 26, 2025, [https://github.com/ory/fosite](https://github.com/ory/fosite)
5. fosite package - github.com/ory/fosite - Go Packages, truy cập vào tháng 10 26, 2025, [https://pkg.go.dev/github.com/ory/fosite](https://pkg.go.dev/github.com/ory/fosite)
6. fosite download | SourceForge.net, truy cập vào tháng 10 26, 2025, [https://sourceforge.net/projects/fosite.mirror/](https://sourceforge.net/projects/fosite.mirror/)
7. Building a RESTful API in Go Using the Gin Framework: A Step-by-Step Tutorial — Part 1/2, truy cập vào tháng 10 26, 2025, [https://medium.com/@godusan/building-a-restful-api-in-go-using-the-gin-framework-a-step-by-step-tutorial-part-1-2-70372ebfa988](https://medium.com/@godusan/building-a-restful-api-in-go-using-the-gin-framework-a-step-by-step-tutorial-part-1-2-70372ebfa988)
8. UUIDv7 Comes to PostgreSQL 18 - Nile Postgres, truy cập vào tháng 10 26, 2025, [https://www.thenile.dev/blog/uuidv7](https://www.thenile.dev/blog/uuidv7)
9. Authentication Token Storage - Redis, truy cập vào tháng 10 26, 2025, [https://redis.io/solutions/authentication-token-storage/](https://redis.io/solutions/authentication-token-storage/)
10. Implementing JWT Authentication with Redis Cache in ASP.NET Core Web API - C\# Corner, truy cập vào tháng 10 26, 2025, [https://www.c-sharpcorner.com/article/implementing-jwt-authentication-with-redis-cache-in-asp-net-core-web-api/](https://www.c-sharpcorner.com/article/implementing-jwt-authentication-with-redis-cache-in-asp-net-core-web-api/)
11. Caching OAuth2 Token using Redis - DEV Community, truy cập vào tháng 10 26, 2025, [https://dev.to/woovi/caching-oauth2-token-using-redis-20d3](https://dev.to/woovi/caching-oauth2-token-using-redis-20d3)
12. Check Access Token every Request with Redis - Stack Overflow, truy cập vào tháng 10 26, 2025, [https://stackoverflow.com/questions/16247319/check-access-token-every-request-with-redis](https://stackoverflow.com/questions/16247319/check-access-token-every-request-with-redis)
13. Revoke Access Using a JWT Blacklist | SuperTokens, truy cập vào tháng 10 26, 2025, [https://supertokens.com/blog/revoking-access-with-a-jwt-blacklist](https://supertokens.com/blog/revoking-access-with-a-jwt-blacklist)
14. Multitenant SaaS Patterns - Azure SQL Database - Microsoft Learn, truy cập vào tháng 10 26, 2025, [https://learn.microsoft.com/en-us/azure/azure-sql/database/saas-tenancy-app-design-patterns?view=azuresql](https://learn.microsoft.com/en-us/azure/azure-sql/database/saas-tenancy-app-design-patterns?view=azuresql)
15. Authorization 101: Multi-tenant RBAC - Aserto, truy cập vào tháng 10 26, 2025, [https://www.aserto.com/blog/authorization-101-multi-tenant-rbac](https://www.aserto.com/blog/authorization-101-multi-tenant-rbac)
16. Multi-Tenant Database Design Patterns 2024 - Daily.dev, truy cập vào tháng 10 26, 2025, [https://daily.dev/blog/multi-tenant-database-design-patterns-2024](https://daily.dev/blog/multi-tenant-database-design-patterns-2024)
17. Designing a Role-Based Access Control (RBAC) System: A ..., truy cập vào tháng 10 26, 2025, [https://medium.com/@07rohit/designing-a-role-based-access-control-rbac-system-a-scalable-approach-441f05168933](https://medium.com/@07rohit/designing-a-role-based-access-control-rbac-system-a-scalable-approach-441f05168933)
18. How do I use storage with postgresql, not memory storage? · Issue \#11 · ory/fosite-example, truy cập vào tháng 10 26, 2025, [https://github.com/ory/fosite-example/issues/11](https://github.com/ory/fosite-example/issues/11)
19. Merge multiple Hydra instances with different system.secrets - Ory, truy cập vào tháng 10 26, 2025, [https://www.ory.sh/docs/hydra/self-hosted/merge-multiple-db-secrets](https://www.ory.sh/docs/hydra/self-hosted/merge-multiple-db-secrets)
20. PostgreSQL UUID Performance: Benchmarking Random (v4) and Time-based (v7) UUIDs, truy cập vào tháng 10 26, 2025, [https://dev.to/umangsinha12/postgresql-uuid-performance-benchmarking-random-v4-and-time-based-v7-uuids-n9b](https://dev.to/umangsinha12/postgresql-uuid-performance-benchmarking-random-v4-and-time-based-v7-uuids-n9b)
21. Releases · gofrs/uuid - GitHub, truy cập vào tháng 10 26, 2025, [https://github.com/gofrs/uuid/releases](https://github.com/gofrs/uuid/releases)
22. uuid package - github.com/gofrs/uuid - Go Packages, truy cập vào tháng 10 26, 2025, [https://pkg.go.dev/github.com/gofrs/uuid](https://pkg.go.dev/github.com/gofrs/uuid)
23. I wrote a UUIDv7 implementation in Go with json.Marshaler and driver.Valuer support : r/golang - Reddit, truy cập vào tháng 10 26, 2025, [https://www.reddit.com/r/golang/comments/1b6wc2b/i_wrote_a_uuidv7_implementation_in_go_with/](https://www.reddit.com/r/golang/comments/1b6wc2b/i_wrote_a_uuidv7_implementation_in_go_with/)
24. OAuth2 & OIDC フレームワークfositeの分析 - [型定義]ストレージ - Zenn, truy cập vào tháng 10 26, 2025, [https://zenn.dev/abcb2/books/analyze-fosite/viewer/storage](https://zenn.dev/abcb2/books/analyze-fosite/viewer/storage)
25. Go 製の認可サーバー、IdP 実装用ライブラリ Fosite - Zenn, truy cập vào tháng 10 26, 2025, [https://zenn.dev/inabajunmr/articles/introduce-ory-fosite](https://zenn.dev/inabajunmr/articles/introduce-ory-fosite)
26. Implementing an oauth 2 server using Go - Charles Muchogo, truy cập vào tháng 10 26, 2025, [https://charles.muchogo.com/posts/articles/authserver_in_go/](https://charles.muchogo.com/posts/articles/authserver_in_go/)
27. github.com/pltfrms/fosite v0.48.0 on Go - Libraries.io - security & maintenance data for open source software, truy cập vào tháng 10 26, 2025, [https://libraries.io/go/github.com%2Fpltfrms%2Ffosite](https://libraries.io/go/github.com%2Fpltfrms%2Ffosite)
28. Using middleware | Gin Web Framework, truy cập vào tháng 10 26, 2025, [https://gin-gonic.com/en/docs/examples/using-middleware/](https://gin-gonic.com/en/docs/examples/using-middleware/)
29. Redis - Fiber Documentation, truy cập vào tháng 10 26, 2025, [https://docs.gofiber.io/storage/redis/](https://docs.gofiber.io/storage/redis/)
30. go-redis guide (Go) | Docs, truy cập vào tháng 10 26, 2025, [https://redis.io/docs/latest/develop/clients/go/](https://redis.io/docs/latest/develop/clients/go/)
31. redis package - github.com/trustbloc/vcs/component/oidc/fosite/redis ..., truy cập vào tháng 10 26, 2025, [https://pkg.go.dev/github.com/trustbloc/vcs/component/oidc/fosite/redis](https://pkg.go.dev/github.com/trustbloc/vcs/component/oidc/fosite/redis)
32. Performance Analysis of an OAuth 2.0-Based Authentication and Authorization System Using a Redis In-memory Database | Request PDF - ResearchGate, truy cập vào tháng 10 26, 2025, [https://www.researchgate.net/publication/394732483_Performance_Analysis_of_an_OAuth_20-Based_Authentication_and_Authorization_System_Using_a_Redis_In-memory_Database](https://www.researchgate.net/publication/394732483_Performance_Analysis_of_an_OAuth_20-Based_Authentication_and_Authorization_System_Using_a_Redis_In-memory_Database)
33. Token Introspection Endpoint - OAuth 2.0 Simplified, truy cập vào tháng 10 26, 2025, [https://www.oauth.com/oauth2-servers/token-introspection-endpoint/](https://www.oauth.com/oauth2-servers/token-introspection-endpoint/)
34. Caching introspection responses - Authlete, truy cập vào tháng 10 26, 2025, [https://www.authlete.com/kb/deployment/performance/caching-introspection-responses/](https://www.authlete.com/kb/deployment/performance/caching-introspection-responses/)
35. OAuth 2.0 and OpenID Connect overview - Okta Developer, truy cập vào tháng 10 26, 2025, [https://developer.okta.com/docs/concepts/oauth-openid/](https://developer.okta.com/docs/concepts/oauth-openid/)
36. Client secrets - SecureAuth Product Docs, truy cập vào tháng 10 26, 2025, [https://docs.secureauth.com/ciam/en/client-secrets.html](https://docs.secureauth.com/ciam/en/client-secrets.html)
37. Best Practices | Authorization Resources - Google for Developers, truy cập vào tháng 10 26, 2025, [https://developers.google.com/identity/protocols/oauth2/resources/best-practices](https://developers.google.com/identity/protocols/oauth2/resources/best-practices)
38. Where to store client_id and client_secret : r/learnprogramming - Reddit, truy cập vào tháng 10 26, 2025, [https://www.reddit.com/r/learnprogramming/comments/18fbqbi/where_to_store_client_id_and_client_secret/](https://www.reddit.com/r/learnprogramming/comments/18fbqbi/where_to_store_client_id_and_client_secret/)
39. Best practices for protecting secrets | Microsoft Learn, truy cập vào tháng 10 26, 2025, [https://learn.microsoft.com/en-us/azure/security/fundamentals/secrets-best-practices](https://learn.microsoft.com/en-us/azure/security/fundamentals/secrets-best-practices)
40. Architecture strategies for protecting application secrets - Microsoft Azure Well-Architected Framework, truy cập vào tháng 10 26, 2025, [https://learn.microsoft.com/en-us/azure/well-architected/security/application-secrets](https://learn.microsoft.com/en-us/azure/well-architected/security/application-secrets)
41. OAuth 2 Refresh Tokens: A Practical Guide - Frontegg, truy cập vào tháng 10 26, 2025, [https://frontegg.com/blog/oauth-2-refresh-tokens](https://frontegg.com/blog/oauth-2-refresh-tokens)
42. Hardening OAuth Tokens in API Security: Token Expiry, Rotation, and Revocation Best Practices | Clutch Events, truy cập vào tháng 10 26, 2025, [https://www.clutchevents.co/resources/hardening-oauth-tokens-in-api-security-token-expiry-rotation-and-revocation-best-practices](https://www.clutchevents.co/resources/hardening-oauth-tokens-in-api-security-token-expiry-rotation-and-revocation-best-practices)
43. Token Best Practices - Auth0, truy cập vào tháng 10 26, 2025, [https://auth0.com/docs/secure/tokens/token-best-practices](https://auth0.com/docs/secure/tokens/token-best-practices)
44. Token Expiry Best Practices | Zuplo Learning Center, truy cập vào tháng 10 26, 2025, [https://zuplo.com/learning-center/token-expiry-best-practices](https://zuplo.com/learning-center/token-expiry-best-practices)
45. Gin Framework Project Structure - Gin for Beginners: Build APIs with Golang Easily, truy cập vào tháng 10 26, 2025, [https://academy.withcodeexample.com/gin-for-beginners-build-apis-with-golang-easily/gin-framework-project-structure](https://academy.withcodeexample.com/gin-for-beginners-build-apis-with-golang-easily/gin-framework-project-structure)
46. Gin gonic project structure : r/golang - Reddit, truy cập vào tháng 10 26, 2025, [https://www.reddit.com/r/golang/comments/17qs1dh/gin_gonic_project_structure/](https://www.reddit.com/r/golang/comments/17qs1dh/gin_gonic_project_structure/)
47. Building a Backend Server with Golang and the Gin Framework | by Mrunmayee Dhapre, truy cập vào tháng 10 26, 2025, [https://medium.com/@mrunmayee.dhapre/building-a-backend-server-with-golang-and-the-gin-framework-ca2cc3eb3721](https://medium.com/@mrunmayee.dhapre/building-a-backend-server-with-golang-and-the-gin-framework-ca2cc3eb3721)
48. sample go/gin project structure. - GitHub, truy cập vào tháng 10 26, 2025, [https://github.com/RezaOptic/gin-project-structure](https://github.com/RezaOptic/gin-project-structure)
