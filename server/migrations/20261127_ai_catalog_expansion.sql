-- 补充 AI 目录条目，保留已有站点和待审投稿。先完成 AI 分类迁移。

BEGIN;
SELECT pg_advisory_xact_lock(82422402);

INSERT INTO "links" ("slug", "category_slug", "name", "url", "desc", "tags", "alias", "sort", "body",
  "lang", "region", "status", "clicks", "views", "check_fails", "user_id", "pinned", "pin_sort", "created_at")
SELECT seed.slug, seed.category_slug, seed.name, seed.url, seed."desc", seed.tags, seed.alias, seed.sort, seed.body,
  CASE WHEN seed.slug IN ('trae-cn', 'qoder-cn', 'tongyi-tingwu') THEN '中文' ELSE '' END,
  CASE WHEN seed.slug IN ('trae-cn', 'qoder-cn', 'tongyi-tingwu') THEN '中国' ELSE '' END,
  'online', 0, 0, 0, 0, FALSE, 0, CURRENT_TIMESTAMP(3)
FROM (
  SELECT 'consensus' AS "slug", 'ai-search' AS "category_slug", 'Consensus' AS "name", 'https://consensus.app/' AS "url", '用研究问题找同行评审论文，浏览不同研究的结论' AS "desc", '论文,学术搜索,文献综述,AI' AS "tags", '' AS "alias", 9 AS "sort", 'Consensus 面向需要从学术论文找证据的读者。可以按自然语言问题、论文标题或 DOI 搜索，再打开结果中列出的论文。对可以回答「是」或「否」的问题，Consensus Meter 会展示工具选中的研究分别支持什么结论。

### 怎么查一个具体问题？

把「远程办公有用吗」改成「混合办公对员工离职率有什么影响」，再看结果里的研究对象、工作模式和观察时间。点开最相关的几篇原文，区分随机实验和观察研究。页面显示的意见分布只对应纳入分析的部分结果，写综述时应标明检索范围。[Consensus 检索指南](https://help.consensus.app/en/articles/9922660-how-to-search-best-practices) · [Consensus Meter 说明](https://help.consensus.app/en/articles/10069920-the-consensus-meter)' AS "body"
  UNION ALL SELECT 'scite', 'ai-search', 'Scite', 'https://scite.ai/', '查论文在后续研究中的引用原文及支持或反驳情况', '论文,引文,学术搜索,AI', '', 10, 'Scite 适合跟踪一篇论文发表后被怎样引用。Smart Citations 提供可跳读的引文片段，帮助区分后续作者是在支持、质疑，还是仅仅提到某项工作。也可以搜索论文正文中的引文。

### 怎样核对一条「被广泛引用」的结论？

用标题或 DOI 找到原论文，打开几条标为反驳或支持的引用，阅读引文前后段落。重点看后续研究讨论的是原结论、测量方法还是另一项实验。引用的分类可以帮你筛选阅读顺序，证据仍应来自引文原文及其研究方法。[Scite 产品说明](https://scite.ai/) · [引文搜索](https://scite.ai/search)'
  UNION ALL SELECT 'recraft', 'ai-ae0faa3203a9', 'Recraft', 'https://www.recraft.ai/', '生成图标与矢量插画，支持导出 SVG 继续编辑', '图像生成,矢量图,SVG,设计,AI', '', 36, 'Recraft 可以生成位图和可继续编辑的矢量图。需要交付网站图标、说明插画或一套同风格视觉元素时，可以在生成前指定使用矢量类型，完成后导出 SVG。

### 怎样做一组能在网页里统一使用的图标？

先确定图标要表达的动作，例如下载、收藏、分享，再写明统一的线条风格、主色和留白。先生成一个代表图标，导出后用矢量编辑器检查轮廓和尺寸，再用同一描述补齐其他图标。最后把它们缩到实际显示大小，检查细线和小形状是否还看得清。[Recraft 矢量图生成说明](https://www.recraft.ai/ai-vector-generator)'
  UNION ALL SELECT 'ideogram', 'ai-ae0faa3203a9', 'Ideogram', 'https://ideogram.ai/', '生成含文字的海报和封面，可在 Studio 编辑识别出的字层', '海报,封面,文字排版,图像生成,AI', '', 37, 'Ideogram 适合制作标题主导的海报、活动封面和社交媒体配图。生成时可以把想放在画面上的短标题写进提示词。Studio 的 Layerize Text 可把识别到的文字转换成可编辑图层，该功能目前以测试版提供。

### 宣传图的标题和日期可以怎么处理？

先用短标题试几版排版，在导出前逐字核对活动时间、地名和品牌拼写。发现错字时，可以试着用 Studio 拆出字层，再替换文字并调整字号、颜色与位置。信息密集的图可以先做背景、最后在排版工具里加完整文案。[Ideogram 字层编辑指南](https://ideogram.ai/features/text-layers/)'
  UNION ALL SELECT 'heygen', 'draw', 'HeyGen', 'https://www.heygen.com/', '把讲稿或幻灯片做成带数字人旁白的讲解视频', '数字人,口播,课件视频,AI', '', 47, 'HeyGen 可用讲稿制作带数字人和旁白的视频，也提供将演示文稿或 PDF 拆成讲解场景的入口。适合做产品操作说明、培训课件或需要固定讲解人出镜的短视频。

### 怎么把产品介绍做成一段口播？

用两三张幻灯片试做，分别讲清场景、操作和结果。每一页只写它要说的重点，缩写或产品名用实际读法试听一遍。预览时核对数字人有没有挡住关键界面、旁白是否跟得上翻页，再补齐剩余页面。采用真人形象或声音时，使用获授权的素材。[HeyGen 演示文稿转视频](https://www.heygen.com/tool/pdf-to-video)'
  UNION ALL SELECT 'pika', 'draw', 'Pika', 'https://pika.art/', '生成短镜头，提供动作迁移、对象替换和镜头延长', '图生视频,短视频,动作迁移,AI', '', 48, 'Pika 的视频工具可以围绕已有画面或短片做单镜头创作。官网列有动作迁移、替换对象、延长视频等独立入口，做素材测试时可以根据手里的输入挑选功能。

### 如何给一张静态插画加动作？

先选主体清楚、四周有留白的图，描述一个主要动作，例如「人物转头，镜头缓慢拉近」。如果有参考动作片段，可以试用动作迁移。先检查一段镜头里的人物边缘、姿态和背景，再决定要延长镜头还是另做转场。发布前查官方说明中的套餐与使用条款。[Pika 视频工具](https://pika.art/) · [Pika 常见问题](https://pika.art/faq)'
  UNION ALL SELECT 'adobe-podcast', 'ai-audio', 'Adobe Podcast', 'https://podcast.adobe.com/', '在线人声降噪，也可录制和编辑播客音频', '播客,人声降噪,录音,配音,AI', '', 29, 'Adobe Podcast 提供网页端录音和音频编辑。拿到已有录音时，可以用 Enhance Speech 减轻背景噪声、改善人声听感。多人远程录制可以从 Studio 开始，录完可下载不同发言人的音轨。

### 有空调声的录音怎么处理？

截一小段同时有人声和噪声的素材，对比处理前后的齿音、呼吸声和句尾。确认声音依然自然再处理全段，并保留原录音，方便对个别片段回退。上传前可在帮助里核对支持的文件格式和当前账户限制。[Adobe Podcast 帮助](https://helpx.adobe.com/podcast/adobe-podcast-faq.html)'
  UNION ALL SELECT 'fish-audio', 'ai-audio', 'Fish Audio', 'https://fish.audio/', '按音色和情绪生成旁白，支持从文字试听语音', '文字转语音,旁白,情绪配音,AI', '', 30, 'Fish Audio 提供文字转语音和声音选择。网页示例列出了强调、低声说话等情绪或表现标签，可以用于比较同一段台词的不同语气。适合为产品演示、故事或分角色脚本先做试音。

### 配一段中文解说时，先试听什么？

选一段有数字、产品名和长句的正文，用同样的文案试听两种音色。把需要停顿的位置用标点分开，再对少数关键句试用情绪标签。合成较长旁白时按场景分段，后期修改一处就不用重做全文。用声音克隆时应取得声音提供者的许可。[Fish Audio 产品与演示](https://fish.audio/)'
  UNION ALL SELECT 'trae-cn', 'ai-code', 'TRAE', 'https://www.trae.cn/', '中文 AI 开发工作台，支持桌面端和网页端', '编程助手,代码编辑器,应用开发,AI', 'TRAE CN', 29, 'TRAE 提供中文 AI 开发工作台，官网有桌面版和网页端入口。它可以把需求拆成开发步骤，也能在开发过程中讨论方案和检查产物。需要理解已有项目时，可以从定位相关文件开始。

### 第一次怎么让它修改一个小功能？

在能运行的项目里选一个小任务，例如给筛选列表加上清空按钮。先说明涉及的页面、点击后的默认状态和验收条件，确认它理解了现有筛选逻辑再动手。完成后查看差异，在浏览器里重走筛选和清空步骤，再决定是否接受改动。[TRAE 官方文档](https://docs.trae.cn/)'
  UNION ALL SELECT 'qoder-cn', 'ai-code', 'Qoder CN', 'https://qoder.cn/', '原通义灵码，提供编程 IDE、JetBrains 插件和终端工具', '编程助手,代码编辑器,JetBrains,AI', '通义灵码', 30, 'Qoder CN 是阿里云原「通义灵码」的现用名称。面向开发的入口包括 Qoder CN IDE、JetBrains 插件和命令行工具。希望继续使用 JetBrains 系列 IDE 的开发者，可以先核对插件支持的编辑器与版本。

### 手头有一个陌生项目，怎么开始？

先请助手从页面、接口或测试任选一个入口解释调用链，并给出对应文件。确认它找对模块后，再让它提出一项范围明确的改动，例如补上缺少的参数校验。检查改动是否沿用现有项目约定，然后跑对应测试。[Qoder CN 系列与入口说明](https://docs.qoder.cn/product-overview/introduction-of-qodercn)'
  UNION ALL SELECT 'replit-agent', 'ai-code', 'Replit Agent', 'https://replit.com/products/agent', '在浏览器里从需求搭建应用原型，支持预览和回退', '应用生成,网站原型,在线开发,AI', 'Replit', 31, 'Replit Agent 在浏览器中根据文字需求建立应用项目。可以在同一工作区查看生成结果、继续提出改动，遇到不合适的变更时可用检查点回退。适合先搭一个能操作的工具原型来验证需求。

### 如何描述一个可验收的应用原型？

以活动报名页为例，先写明表单字段、提交后的提示、重复报名的处理规则，再给几条模拟数据。打开预览实际提交一遍，同时试试空表单和重复提交。确定基本流程成立后再接真实数据和发布所需配置。[Replit Agent 文档](https://docs.replit.com/features/agent/overview)'
  UNION ALL SELECT 'tongyi-tingwu', 'ai-office', '通义听悟', 'https://tingwu.aliyun.com/', '实时转写或上传录音，整理会议纪要和待办', '会议纪要,录音转文字,学习笔记,AI', '听悟', 38, '通义听悟是阿里的音视频整理工具，可实时把发言转成文字，也能对已录制的音视频做转写。会议和课程内容可以按说话人或章节回看，便于定位关键信息、整理待办。

### 会后怎么把录音整理成可发的纪要？

先转写录音，按章节找到涉及决定和分工的片段。用时间点回听姓名、金额、日期等容易转错的内容。再把待办写成「谁、何时、交付什么」三项，留空的部分找当事人确认。上传含内部信息的录音前，确认组织允许使用该服务。[通义听悟官网](https://tingwu.aliyun.com/) · [阿里云功能说明](https://help.aliyun.com/zh/tingwu/benefits)'
  UNION ALL SELECT 'napkin-ai', 'ai-office', 'Napkin AI', 'https://www.napkin.ai/', '把说明文字转成可编辑的流程图、思维导图或图表', '流程图,思维导图,汇报配图,AI', 'Napkin', 39, 'Napkin AI 可以从已有文字生成流程图、思维导图等可视化内容，再调整节点和版式。它适合把报告里难读的长段落变成一张说明图，也能做演示文稿里的业务流程配图。

### 怎样把审批制度写成一张流程图？

先把原文改写成「谁发起、谁审核、通过后到哪里、退回后怎么办」，每一步单独一行。生成图后沿通过和退回两条路径分别走一遍，检查箭头方向和职责。把节点文字缩到能在演示页读清的长度，再导出到文档中。[Napkin AI 流程图说明](https://www.napkin.ai/flowcharts/)'
  UNION ALL SELECT 'make-ai-agents', 'ai-6bf6d3f3d626', 'Make AI Agents', 'https://www.make.com/ai-agents', '在可视化流程中接入 AI 分类、提取和应用动作', '工作流,自动化,智能体,AI', 'Make', 27, 'Make 可以把不同应用连在同一个可视化场景中。AI Agents 适合处理需要理解非结构化内容的步骤，例如给反馈摘要、分类，再交给场景里的其他模块继续处理。场景能展示各模块的执行过程。

### 第一条流程可以怎么验证？

用一张测试表和几条示例反馈搭起「读取一行、提取主题、记录分类」三个步骤。先手动运行，检查重复数据和分类失败时落到了哪里。等输出格式稳定后，再连接真实的表单或通知工具，并保留需要人工确认的分支。[Make AI Agents 产品说明](https://www.make.com/ai-agents)'
  UNION ALL SELECT 'activepieces', 'ai-6bf6d3f3d626', 'Activepieces', 'https://www.activepieces.com/', '连接工作流、AI 智能体和数据表，提供云端与自托管方式', '工作流,自动化,自托管,智能体,AI', '', 28, 'Activepieces 把触发器、应用连接器、数据表和 AI 智能体放在一个自动化环境中。可以使用托管版本，也有面向团队自托管的安装文档。适合做按时执行或收到数据后自动处理的流程。

### 怎样做一个能检查结果的自动化流程？

从一个可反复测试的场景开始，例如读取表单示例，把反馈摘要写入专用测试表。为每条输入保留原始记录编号，重复执行时按编号核对是否写出重复行。需要连接团队工具时，先用测试账号验证连接和失败记录，再在线上启用。[Activepieces 官方入门](https://www.activepieces.com/docs/getting-started/introduction)'
) AS seed
JOIN "categories" AS category ON category.slug = seed.category_slug AND category.parent_slug = 'ai'
WHERE NOT EXISTS (
  SELECT 1 FROM "links" AS existing
  WHERE existing.slug = seed.slug OR existing.name = seed.name
    OR existing.alias = seed.name OR (seed.alias <> '' AND existing.name = seed.alias)
    OR TRIM(TRAILING '/' FROM existing.url) = TRIM(TRAILING '/' FROM seed.url)
);

COMMIT;
