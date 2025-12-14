#Phase 2: Data Storage 💾
    เป้าหมาย: จำงานที่เคยเจอแล้ว เพื่อไม่แจ้งซ้ำ
    Tasks:
    ✓ เลือก storage (แนะนำ jobs.json สำหรับเริ่มต้น)
    ✓ สร้าง functions:
    - LoadJobs() - อ่านข้อมูลเก่า
    - SaveJobs() - บันทึกข้อมูลใหม่
    - IsNewJob() - เช็คว่างานนี้เคยเห็นไหม
    ✓ เก็บ job ID หรือ URL เป็น unique key
    ✓ Test: scrape 2 ครั้ง ต้องไม่มีงานเดิมซ้ำ
    Output: ระบบจำงานที่เคยเห็นแล้วได้

#Phase 3: Notification System 🔔
    เป้าหมาย: ส่งข้อความแจ้งเตือนเมื่อเจองานใหม่
    Tasks:
    ✓ สร้าง Discord webhook หรือ LINE Notify token
    ✓ เขียน function SendNotification(job Job)
    ✓ Format ข้อความให้สวยงาม:
    🔥 New Internship!
    📌 Position: [ตำแหน่ง]
    🏢 Company: [บริษัท]
    🔗 Link: [URL]
    ✓ Test ส่งข้อความทดสอบ
    ✓ เชื่อมกับ Phase 2: เจองานใหม่ → ส่งทันที
    Output: ได้รับ notification ใน Discord/LINE เมื่อมีงานใหม่

#Phase 4: Scheduler & Automation ⏰
    เป้าหมาย: รันอัตโนมัติทุก 10-30 นาที
    Option A: ใช้ Cron ใน Go
    ✓ ติดตั้ง robfig/cron
    ✓ ตั้ง schedule: "*/15 * * * *" (ทุก 15 นาที)
    ✓ Wrap scraping logic เป็น function
    ✓ เพิ่ม logging เพื่อ debug
    Option B: ใช้ GitHub Actions
    ✓ สร้าง .github/workflows/scraper.yml
    ✓ ตั้ง cron schedule
    ✓ Run script ใน GitHub
    ✓ เก็บ jobs.json ใน artifacts หรือ gist
    Output: Bot รันเองอัตโนมัติโดยไม่ต้องเปิดเครื่อง

#Phase 5: Error Handling & Resilience 🛡️
    เป้าหมาย: ทำให้ระบบไม่พังง่าย ๆ
    Tasks:
    ✓ Retry logic: ถ้า request fail → retry 3 ครั้ง
    ✓ Timeout handling (5-10 วินาที)
    ✓ HTML structure changes detection
    ✓ Error notification (ส่งแจ้งว่า bot พัง)
    ✓ Logging: เก็บ log file เพื่อ debug
    ✓ Rate limiting: ใส่ delay ระหว่าง requests