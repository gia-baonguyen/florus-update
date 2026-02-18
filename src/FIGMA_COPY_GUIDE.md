# 🎨 Hướng dẫn Copy Design sang Figma

## 🎯 TL;DR - Nhanh Gọn

**Vấn đề:** Banner không hiển thị khi copy sang Figma  
**Giải pháp:** ✅ Đã fix! Giờ banner dùng gradient + emoji, sẽ hiển thị đầy đủ khi copy

### Copy ngay bây giờ:
1. Mở app: `npm run dev`
2. Chọn phần muốn copy (Ctrl/Cmd + A để chọn tất cả)
3. Copy: Ctrl/Cmd + C
4. Mở Figma → Paste: Ctrl/Cmd + V
5. ✅ Done! Banner + tất cả sản phẩm sẽ hiển thị với gradient đẹp mắt

---

## ✅ Đã giải quyết vấn đề ảnh bị mất khi copy

Trước đây khi bạn copy từ browser sang Figma, các ảnh từ URL bên ngoài (Unsplash) sẽ **bị mất (trắng trơn)**.

**Giải pháp mới:**
- ✅ Đã thay thế tất cả ảnh thật bằng **gradient backgrounds + emoji** đại diện
- ✅ Banner carousel cũng đã được update với gradient + decorative emojis
- ✅ Khi copy sang Figma, bạn sẽ thấy **màu sắc gradient đẹp mắt** với icon emoji
- ✅ Gradient được thiết kế phù hợp với từng loại sản phẩm beauty (son môi = coral pink, serum = orange, etc.)
- ✅ Layout và spacing được giữ nguyên hoàn hảo

## 🚀 Chế độ Static Banner cho Figma Copy

**Banner carousel có animation nên khi copy có thể bị mất nội dung.**

### Cách bật Static Banner Mode:

1. Mở file `/App.tsx`
2. Tìm dòng:
   ```typescript
   const useStaticBanner = false; // Change to true for Figma copy mode
   ```
3. Đổi thành:
   ```typescript
   const useStaticBanner = true; // Static mode - perfect for Figma copy
   ```
4. Save và refresh browser
5. Giờ banner sẽ hiển thị tĩnh (không slide), dễ copy sang Figma hơn

### Khi nào dùng Static Mode:
- ✅ Khi muốn copy design sang Figma
- ✅ Khi chụp screenshots cho báo cáo
- ✅ Khi cần hiển thị đầy đủ layout không bị mất element

### Khi nào dùng Animated Mode (default):
- ✅ Khi demo ứng dụng thực tế
- ✅ Khi present cho khách hàng/giảng viên
- ✅ Khi quay video demo

## 📋 Cách copy từ Browser sang Figma

### Bước 1: Mở ứng dụng trong browser
```bash
npm run dev
```
Truy cập: http://localhost:5173

### Bước 2: Chọn vùng muốn copy
- Dùng chuột chọn toàn bộ hoặc một phần màn hình
- Ví dụ: Chọn toàn bộ trang Home, hoặc chỉ một Product Card

### Bước 3: Copy
- **Windows/Linux:** Ctrl + C
- **Mac:** Cmd + C

### Bước 4: Paste vào Figma
1. Mở Figma Desktop hoặc Web
2. Tạo một Frame mới hoặc mở file có sẵn
3. Paste:
   - **Windows/Linux:** Ctrl + V
   - **Mac:** Cmd + V

### Kết quả:
✅ Layout hoàn chỉnh với spacing chính xác
✅ Màu sắc gradient đẹp mắt thay cho ảnh
✅ Typography (Playfair Display Serif + Inter Sans-serif)
✅ Border radius, shadows, và tất cả styling

## 🎨 Màu Gradient cho từng sản phẩm

### Makeup Products:
- **Lipstick (💄)**: Coral Pink gradient (#FF6B6B → #FFB4B4)
- **Eyeshadow (🎨)**: Nude/Brown gradient (#D4A574 → #F5E3D3)
- **Mascara (👁️)**: Black/Gray gradient (#2D2D2D → #8C8C8C)
- **Blush (🍑)**: Peachy gradient (#FFAA85 → #FFD6C2)
- **Nail Polish (💅)**: Pastel gradient (#FFB3E6 → #B8C9FF)

### Skincare Products:
- **Face Cream (✨)**: Rose Gold gradient (#E8B4A8 → #FFEAE5)
- **Vitamin C Serum (🍊)**: Orange gradient (#FFB84D → #FFF0D9)
- **Sheet Mask (😊)**: Green/Peach gradient (#A8E6CF → #FFD3B6)
- **Night Cream (🌙)**: Purple gradient (#9B84C7 → #E8DFF5)
- **Cleansing Water (💧)**: Blue gradient (#89CFF0 → #E8F7FF)

### Fragrance:
- **Perfume (🌸)**: Pink gradient (#FF85B3 → #FFE0F0)

## 🔄 Nếu muốn thay thế ảnh thật trong Figma

Sau khi copy sang Figma, bạn có thể:

1. **Giữ nguyên gradient + emoji** (Recommended cho mockup/presentation)
   - Trông hiện đại, minimalist
   - Phù hợp với concept "Florus Beauty"
   - Không bị vấn đề copyright

2. **Replace với ảnh thật:**
   - Click vào emoji placeholder trong Figma
   - Chọn "Fill" > "Image"
   - Upload ảnh sản phẩm thật từ máy tính
   - Adjust: Fill type = "Fill" hoặc "Fit"

## 📱 Các màn hình đã sẵn sàng để copy

### 1. **Home Page (Trang chủ)**
- ✅ Cold-start view: Popular products
- ✅ Warm-start view: Personalized recommendations
- ✅ Banner carousel
- ✅ Category grid

### 2. **Product Detail Page**
- ✅ Large product image với thumbnail gallery
- ✅ Content-based recommendations section
- ✅ Collaborative filtering (Neo4j) section
- ✅ AI insight box

### 3. **Cart Page**
- ✅ Cart items list
- ✅ Order summary
- ✅ Cross-sell recommendations widget
- ✅ Free shipping threshold

### 4. **Mobile Navigation**
- ✅ Bottom tab bar
- ✅ Sticky buy button on product detail

## 🎯 Tips cho việc copy sang Figma

### ✅ DO:
- Copy từng section một nếu file quá lớn
- Resize frame trong Figma theo thiết bị: Mobile (375px), Desktop (1440px)
- Dùng Auto Layout trong Figma để dễ chỉnh sửa

### ❌ DON'T:
- Copy khi browser zoom không phải 100%
- Copy khi có scroll bar đang hiện
- Copy overlay/modal đang mở (đóng hết trước khi copy)

## 🎨 Design System được giữ nguyên

Khi copy sang Figma, bạn sẽ có đầy đủ:

### Colors:
- Primary: #FF6B6B (Coral Pink)
- Secondary: #1FAB89 (Teal Blue)
- Rose Gold: #E8B4A8
- Backgrounds: #FFFFFF, #F5F5F5, #FFF9F7

### Typography:
- Headings: Playfair Display (Serif) - 600 weight
- Body: Inter (Sans-serif) - 400/500 weight
- Sizes được giữ nguyên từ CSS

### Spacing:
- Rounded corners: 12px-16px (rounded-xl, rounded-2xl)
- Padding/margin: Tailwind scale (4px, 8px, 12px, 16px, 24px)
- Shadows: Soft shadows với opacity thấp

## 🚀 Workflow đề xuất cho báo cáo

1. **Chụp screenshots của 4 màn hình chính**
   - Dùng browser DevTools để chuyển qua Mobile/Desktop view
   - Cmd/Ctrl + Shift + P → Capture screenshot

2. **Hoặc copy trực tiếp sang Figma**
   - Tạo 4 frames trong Figma tương ứng 4 màn hình
   - Copy từng màn hình vào từng frame
   - Export as PNG/PDF cho báo cáo

3. **Annotations trong Figma**
   - Thêm text boxes giải thích các tính năng AI
   - Point out: "Cold-start Engine", "Warm-start AI", "Knowledge Graph", "Cross-sell"
   - Highlight các badges: "AI Pick ✨", "Trending 🔥", etc.

## 💡 Lưu ý quan trọng

**Gradient + Emoji thay vì ảnh thật là có chủ đích:**

✅ **Ưu điểm:**
- Không vi phạm copyright khi dùng ảnh Unsplash
- Copy sang Figma không bị mất layout
- Trông professional và minimalist
- Phù hợp với concept "Clean & Tech-focused"
- Dễ dàng customize màu sắc trong Figma

✅ **Cho báo cáo kỹ thuật:**
- Tập trung vào **công nghệ AI/ML** chứ không phải ảnh sản phẩm
- Visual cues rõ ràng về recommendation engines
- Dễ annotate và giải thích flow

---

**Chúc bạn thành công với đồ án! 🎉**

Nếu cần adjust màu gradient hoặc emoji nào, hãy báo để tôi customize thêm.