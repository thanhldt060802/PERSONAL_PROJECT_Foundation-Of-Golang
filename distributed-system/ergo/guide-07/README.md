<h1><strong><i>Các thí nghiệm với Ergo trong xử lý đồng thời theo mô hình Actor</i></strong></h1>

<h3><strong>Ví dụ 01:</strong></h3>
<div><strong>2 Sender Call() đồng bộ tới 1 Receiver trên cùng một node.</strong></div>
<ul>
    <li>Khi 2 Sender Call() đồng thời tới 1 Receiver thì tuỳ vào Receiver nhận tín hiệu của Sender nào trước, nó xử lý xong và trả kết quả về cho Sender đó trước rồi mới tới Sender còn lại, trong lúc Receiver xử lý thì Sender phải chờ Receiver trả về kết quả.</li>
</ul>

<h3><strong>Ví dụ 02:</strong></h3>
<div><strong>2 Sender Send() bất đồng bộ tới 1 Receiver trên cùng một node.</strong></div>
<ul>
    <li>Khi 2 Sender Call() đồng thời tới 1 Receiver thì tuỳ vào Receiver nhận tín hiệu của Sender nào trước, nó xử lý xong và trả kết quả về cho Sender đó trước rồi mới tới Sender còn lại, trong lúc Receiver xử lý thì Sender không phải chờ Receiver vì Receiver sẽ không trả về kết quả.</li>
</ul>

<h3><strong>Ví dụ 03:</strong></h3>
<div><strong>2 Sender trên node1@localhost Call() đồng bộ tới 2 Receiver lần lượt là local trên cùng node và remote trên node2@localhost.</strong></div>
<ul>
    <li>Khi tạo node1 và node2 cần cho chúng sử dụng chung một Cookie (đóng vai trò bảo mật liên lạc giữ 2 node) và đăng ký Network Message cho các Message dùng để Sender gửi quan Remote Receiver.</li>
</ul>

<h3><strong>Ví dụ 04:</strong></h3>
<div><strong>2 Sender trên node1@localhost Send() bất đồng bộ tới 2 Receiver lần lượt là local trên cùng node và remote trên node2@localhost.</strong></div>
<ul>
    <li>Khi tạo node1 và node2 cần cho chúng sử dụng chung một Cookie (đóng vai trò bảo mật liên lạc giữ 2 node) và đăng ký Network Message cho các Message dùng để Sender gửi quan Remote Receiver.</li>
</ul>

<h3><strong>Ví dụ 05:</strong></h3>
<div><strong>Sử dụng Observe để quan sát trên các node đang vận hành có liên quan.</strong></div>
<ul>
    <li>Quan sát các Actor đang chạy trên node và ở node khác nếu có các tương tác Remote.</li>
    <li>Thống kê số liệu về các Actor đang vận hành cũng như các thành phần quan trọng khác trong mô hình Actor.</li>
</ul>

<h3><strong>Ví dụ 06:</strong></h3>
<div><strong>Sử dụng Supervisor để quản lý, giám sát (và có thể điều phối cho) N Worker.</strong></div>
<ul>
    <li>Các chiến lượt Restart:<br>
        <ul>
            <li>Permanent: Restart lại Actor khi bị lỗi (kể cả exit signal).</li>
            <li>Transient: Restart lại Actor khi bị lỗi (bỏ qua exit signal).</li>
            <li>Temporary: Không Restart lại Actor khi bị lỗi.</li>
        </ul>
    </li>
    <li>Các loại Supervisor:<br>
        <ul>
            <li>OneForOne: Khi một Actor bị Restart, không ảnh tới các Actor còn lại.</li>
            <li>AllForOne: Khi một Actor bị Restart, tất cả các Actor còn lại cũng bị Restart.</li>
            <li>RestForOne: Khi một Actor bị Restart, các Actor phía sau nó sẽ bị Restart.</li>
        </ul>
    </li>
</ul>

<h3><strong>Ví dụ 07:</strong></h3>
<div><strong>Sử dụng Pool để quản lý N Worker và điều phối cho 1 Worker trong nhóm.</strong></div>
<ul>
    <li>Khi Pool có N Actor thì việc Call() hay Send() tới Pool sẽ được nó điều hướng tương ứng tới HandleCall() và HandleMessage() của một Actor nào đó trong Pool.</li>
    <li>Với Call() là đồng bộ nhưng dùng với Pool thì ta có thể Call() và đồng thời chờ Pool trả về kết quả với số lượng ứng với số lượng Actor trong Pool.</li>
</ul>

<h3><strong>Ví dụ 08:</strong></h3>
<span>Sử dụng TCP để xử lý các gói tin.</span>
<ul>

</ul>

<h3><strong>Ví dụ 09:</strong></h3>
<span>Sử dụng WebWorker để tạo một trình xử lý các yêu cầu HTTP/HTTPS từ client.</span>
<ul>

</ul>

<h3><strong>Ví dụ 10:</strong></h3>
<span>Sử dụng WebSocket để tạo môi trường real-time cho các client.</span>
<ul>

</ul>

<h3><strong>Ví dụ 11:</strong></h3>
<div><strong>Thiết lập FSM trong mô hình Actor.</strong></div>
<ul>
    <li>Khi một Actor muốn xử lý một công đoạn nào đó sẽ cần dữa vào trạng thái hiện tại để có thể xử lý sự kiện và chuyển sang trạng thái tiếp theo.</li>
</ul>

<h3><strong>Ví dụ 12:</strong></h3>
<div><strong>Tích hợp mô hình Actor vào Service API.</strong></div>
<ul>
    <li>Nhúng các thành phần của node vào trong Service Layer để có thể xử lý tương tác với Actor trong Service Layer.</li>
    <li>Nhứng các thành phần liên quan đến database vào mô hình Actor để có thể thực hiện tương tác với database trong mô hình Actor.</li>
    <li>Quản lý các Actor một cách linh hoạt, đảm bảo hiệu suất.</li>
    <li>Thống kê được thông tin cơ bản của các Actor hoạt động trên node.</li>
</ul>

<h3><strong>Ví dụ 13:</strong></h3>
<div><strong>Xây dựng mô hình Actor tương tác trên cùng một Data Source.</strong></div>
<ul>
    <li>Các Worker chạy đồng thời và truy cập đến cùng một Data Source.</li>
    <li>Sử dụng sync.Mutex để thực hiện cơ chế khoá/mở khoá đảm bảo đồng bộ hoá tương tác với Data Source trong môi trường truy cập đồng thời.</li>
</ul>

<h3><strong>Ví dụ 14:</strong></h3>
<div><strong>Xây dựng mô hình Actor chờ nhận và xử lý theo các Event từ một điểm truy cập.</strong></div>
<ul>
    <li>Event Dispatcher đóng vai trò là điểm truy cập và phản hồi theo các Event.</li>
    <li>Các Worker gửi Request đến Event Dispatcher và lắng nghe các Event được gửi lại từ phía Event Dispatcher và xử lý theo từng Event nhận được.</li>
</ul>

<br><br>

<h1><strong><i>Các demo ứng dụng của Ergo trong xử lý đồng thời theo mô hình Actor</i></strong></h1>

<h3><strong>Demo 01:</strong></h3>
<div><strong>Mô hình Actor cho phép chạy đồng thời các Task. Biết Task được lấy trên PostgreSQL có sẵn.</strong></div>
<ul>
    <li>Thực hiện xử lý Task đồng thời và cập nhật trạng thái Task kịp thời khi xảy ra lỗi.</li>
    <li>Quản lý các Actor một cách linh hoạt, đảm bảo hiệu suất.</li>
    <li>Thống kê được thông tin cơ bản của các Actor hoạt động trên node.</li>
</ul>
