-- 适用于已经有内容的开物录数据库。请先选择目标数据库，再导入此文件。

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `category_guides` (
  `category_slug` varchar(64),
  `title` varchar(128),
  `intro` text,
  `home_label` varchar(64),
  PRIMARY KEY (`category_slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `history_events` (
  `id` bigint unsigned AUTO_INCREMENT,
  `month` bigint NOT NULL,
  `day` bigint NOT NULL,
  `year` bigint NOT NULL,
  `text` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `ux_history_event` (`month`,`day`,`year`,`text`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

START TRANSACTION;

INSERT INTO `category_guides` (`category_slug`, `title`, `intro`, `home_label`) VALUES
  ('chat', 'AI 对话助手', '写提纲、读长文、查资料和解释代码，可以从 AI 对话助手开始。读文档前先确认能上传的格式和大小，查资料时优先选能列出来源的工具，并打开原文核对。', 'AI 对话'),
  ('design', '设计素材与在线设计', '找图标、配色、作品参考或在线协作工具，可以按项目需要挑选。下载素材前核对单个作品的授权和署名要求，交付前确认能否导出所需格式。', '设计素材'),
  ('draw', 'AI 绘画与视频生成', '做插画、商品图或短视频时，先按文生图、图生图或图生视频挑工具。需要交付成品时，生成前先看导出尺寸、水印和商用条款。', '绘图与视频'),
  ('learn', '学习网站与公开课', '想系统学习可从大学公开课入手，查 Web API 可看开发文档，刚开始编程可跟着短示例练习。选课时先看先修要求、授课语言和课后练习。', '学习资源'),
  ('query', '生活实用查询', '按任务找查询入口，出行可查车次、航班和天气，寄件可查物流，核对企业信息可查公示平台。涉及办事或资质核验时，以对应官方入口为准。', '实用查询'),
  ('webtool', '在线网页工具', '临时压缩图片、格式化 JSON、换算时间戳或生成二维码，可以在浏览器里完成。处理客户文件、密钥和日志时，先确认工具是否会上传数据。', '网页工具')
ON DUPLICATE KEY UPDATE `category_slug` = `category_guides`.`category_slug`;

INSERT INTO `history_events` (`month`, `day`, `year`, `text`) VALUES
  (1, 1, 1863, '美国《解放奴隶宣言》生效。'),
  (1, 24, 1984, '苹果麦金塔个人电脑开始发售。'),
  (2, 11, 1847, '爱迪生出生。'),
  (3, 7, 1876, '贝尔获得电话专利。'),
  (3, 14, 1879, '爱因斯坦出生。'),
  (4, 12, 1961, '加加林完成人类首次载人航天。'),
  (4, 23, 1995, '联合国教科文组织把 4 月 23 日定为世界读书日。'),
  (5, 4, 1919, '五四运动爆发。'),
  (6, 5, 1972, '联合国把 6 月 5 日定为世界环境日。'),
  (7, 1, 1997, '香港回归中国。'),
  (7, 20, 1969, '阿波罗 11 号登月。'),
  (8, 8, 2008, '北京奥运会开幕。'),
  (9, 10, 1985, '中国第一个教师节。'),
  (10, 1, 1949, '中华人民共和国成立，开国大典举行。'),
  (10, 3, 1990, '两德正式统一。'),
  (10, 3, 1929, '塞尔维亚-克罗地亚-斯洛文尼亚王国改名为南斯拉夫王国。'),
  (11, 9, 1989, '柏林墙开放。'),
  (12, 17, 1903, '莱特兄弟完成首次有动力、可持续的飞行。')
ON DUPLICATE KEY UPDATE `id` = `history_events`.`id`;

-- 仅替换与旧简述和旧正文完全匹配的记录，保留后台编辑或清空过的内容。

UPDATE `links`
SET `desc` = 'AI 视频与图片创作，支持文生视频、图生视频和短剧分镜',
    `body` = '小云雀是抖音旗下的网页端 AI 创作工具。可以输入文字描述或上传参考图片生成视频，也能生成图片。已有剧本的创作者可以用短剧功能整理角色、场景和分镜。

### 小云雀怎么开始做视频？

做一条短片，可以先描述画面和动作，或上传图片作为参考。做短剧或漫剧时，可以粘贴剧本，也可以导入 TXT 或 DOCX 文件，再按场景和分镜继续制作。

参考资料：[小云雀官网](https://xyq.jianying.com/) · [小云雀 Web 官方产品手册](https://xyq.jianying.com/tutorials/web-manual)。'
WHERE `slug` = 'xyq-jianying'
  AND `name` = '小云雀'
  AND `url` = 'https://xyq.jianying.com'
  AND CAST(`desc` AS BINARY) = '短片和故事'
  AND CAST(`body` AS BINARY) = '小云雀是抖音旗下的创作工具，用一段话生成短片和画面，网页在剪映这套域名下。登录一般走抖音或字节的账号。

生成结果里的人脸和商标要自己看一遍。发布前确认这次还有没有可用次数。';

UPDATE `links`
SET `desc` = 'AI 动画制作工具，支持剧本、角色设计、分镜和多镜头成片',
    `body` = 'OiiOii 可以从文字构思、已有剧本或角色图开始制作动画。它的 Story Video 工作流适合做有多个镜头的故事，例如同一角色在不同场景里连续出场。只想让一张插画动起来，可以直接用图生视频。

### OiiOii 怎么做一段故事动画？

进入官网的 Story Video，输入故事梗概，或上传已有剧本。先确认剧本和画风，再检查角色与场景。在分镜阶段，可以在同一段对话里要求调整动作、构图或镜头顺序，确认后再生成片段并合成。第一次试用可以先做三个镜头，例如「猫走进雨巷、发现湿透的纸船、把纸船推回水沟」。先看三个镜头能否讲清楚故事，再生成视频。[官方动画制作指南](https://www.oiioii.ai/en/ai-animation-generator)

### 同一个角色怎么在多个镜头里出现？

把满意的角色设定保存为素材。做后续故事时，在创作对话里用「@」引用该素材，也可以用同样的方法复用场景。这个功能让后续生成有同一份角色和场景参考，结果仍会随模型和镜头变化。生成前可以先在分镜里核对服装、道具和场景是否接得上。[角色与场景复用说明](https://www.oiioii.ai/en/ai-animation-generator)

### 可以免费试用吗？

官方说明分镜和视频生成都会消耗积分，免费积分足够时可以用来试做。不同模型的积分消耗不同，做多镜头动画前，先看所选模型的消耗和当前余额。[官方积分说明](https://app.oiioii.art/features/video-generation) · [会员与积分价格](https://www.oiioii.ai/price)'
WHERE `slug` = 'oiioii'
  AND `name` = 'OiiOii'
  AND `url` = 'https://www.oiioii.ai'
  AND CAST(`desc` AS BINARY) = '动画分镜'
  AND CAST(`body` AS BINARY) = 'OiiOii 用来做动画分镜，把一段想法拆成镜头。先写清角色和故事，再看它排出的画面。

镜头只是草稿。人物前后不一致时，锁定角色再生成下一镜。';

UPDATE `links`
SET `desc` = '多模型 AI 聊天平台，可选 GPT、Claude 和 Gemini',
    `body` = 'Poe 是 Quora 提供的多模型 AI 聊天平台，同一个账号可以体验 GPT、Claude、Gemini 等模型，也可以使用官方或用户创建的机器人。它适合用同一份任务对比几个模型的回答，或试用专门做图片、视频和写作的机器人。

### Poe 怎么选模型？

打开 [Poe 机器人目录](https://poe.com/explore)，搜索模型名，点进机器人就能看它适合做什么。想比较 GPT 和 Claude 的中文改写效果，可以分别新建对话，把同一段文字和同样的要求发过去，再对照事实有没有被改动、语气是否自然。新开对话可以避免之前的聊天内容影响比较。

### Poe 免费吗？积分怎么扣？

免费账号会收到每日积分，未使用的每日积分不会累积。各机器人扣分规则不同，有的按一次生成扣分，有的会随输入和输出长度变化。做长文总结或生成视频前，先点机器人的名称看具体规则，再决定用哪个模型。订阅提供更多积分和部分模型的使用权限。具体额度可查 [Poe 官方积分与机器人说明](https://help.poe.com/hc/en-us/articles/19944206309524-Poe-FAQs)。

### 怎么创建自己的机器人？

从 [创建机器人](https://poe.com/create_bot) 选择基础模型，写下它该怎么回答。例如，建一个中英文邮件修改助手，要求每次先列出容易误解的句子，再给出正式和简短两版。试用阶段可以先设为私有。更多创建方式见 [Poe 创作者文档](https://creator.poe.com/docs)。'
WHERE `slug` = 'poe'
  AND `name` = 'Poe'
  AND `url` = 'https://poe.com'
  AND CAST(`desc` AS BINARY) = '多个模型对话'
  AND CAST(`body` AS BINARY) = 'Poe，多个模型对话。先看它是不是你要找的那一个，再打开。

页面改版、要登录或分地区开放时，以网站自己的说明为准。';

UPDATE `links`
SET `desc` = 'Google AI 助手，支持对话与文件分析',
    `body` = 'Gemini 是 Google 的 AI 助手，可在网页上起草文字、改写、分析文件。比如把两份公开的 PDF 放在同一轮对话里，先让它分别概括，再按关注的几个问题比较差异。

### Gemini 怎么读 PDF 和表格？

在对话框旁添加文件，再说明要查什么。读报告时，可以要求按「结论、原文依据、页码」输出。看表格时，先请它写清用了哪些列、单位和筛选条件，再把关键数字回到原表复核。支持的文件和上传限额见 [Gemini 文件分析帮助](https://support.google.com/gemini/answer/14903178)。

### Gemini 和 NotebookLM 怎么选？

临时改一段文字或问一份附件，可以先用 Gemini 对话。若要围绕一组论文或课件反复提问、核对出处，适合把这些资料放进 [Gemini Notebook（原 NotebookLM）](/site/notebooklm-google)，日后继续在同一个笔记本里查。'
WHERE `slug` = 'gemini-google'
  AND `name` = 'Gemini'
  AND `url` = 'https://gemini.google.com'
  AND CAST(`desc` AS BINARY) = '对话'
  AND CAST(`body` AS BINARY) = 'Gemini，对话。先看它是不是你要找的那一个，再打开。

页面改版、要登录或分地区开放时，以网站自己的说明为准。';

UPDATE `links`
SET `desc` = '微软个人版 AI 助手，支持网页问答与文件分析',
    `body` = 'Microsoft Copilot 的个人网页入口是 copilot.microsoft.com，可以用来提问、起草内容，也可以上传支持的文件或图片让它分析。短任务可以直接在网页上开始，登录个人 Microsoft 账号后还能使用聊天记录等功能。

### 网页版 Copilot 可以怎么用？

拿一份公开的活动介绍试试，让它按「时间、地点、报名截止、费用」提取，缺失的项目标为未说明。看网页总结时，可以在 Edge 中用 Copilot 整理当前页面，再回原文核对数字和关键条款。功能入口和文件支持见 [Microsoft Copilot 功能说明](https://www.microsoft.com/en/microsoft-copilot/for-individuals/features)。

### Copilot 和 Microsoft 365 Copilot 有什么区别？

个人网页版适合日常问答和写作。想在 Word 或 Excel 里直接用 Copilot，要看 Microsoft 365 订阅及账号权限。工作或学校账号由组织管理，能否使用及能访问哪些资料取决于组织配置。可参照 [微软的版本说明](https://support.microsoft.com/en-us/microsoft-365-copilot/what-s-the-difference-between-microsoft-copilot-free-and-copilot-in-microsoft-365)。'
WHERE `slug` = 'copilot-microsoft'
  AND `name` = 'Copilot'
  AND `url` = 'https://copilot.microsoft.com'
  AND CAST(`desc` AS BINARY) = '对话'
  AND CAST(`body` AS BINARY) = 'Copilot，对话。先看它是不是你要找的那一个，再打开。

页面改版、要登录或分地区开放时，以网站自己的说明为准。';

UPDATE `links`
SET `desc` = 'Gemini Notebook，按来源做文档问答与音频概览',
    `body` = 'NotebookLM 已更名为 Gemini Notebook，是 Google 的资料研究工具。把 PDF、网页或支持的 Google Drive 文件放进同一个笔记本，就可以围绕这组资料问问题、查看回答引用的原文，还可以生成音频概览和复习材料。

### NotebookLM 怎么用来读论文？

在 [Gemini Notebook](https://notebook.google.com/) 新建笔记本，把同一主题的几篇论文分别添加为来源。可以问「请按论文列出样本量、研究对象和结论，并给出原文引用」。随后点引用回到对应资料。只想核对其中一篇时，在来源面板只选那篇。操作可查 [Google 的来源管理说明](https://support.google.com/gemininotebook/answer/16215270) 和 [对话与引用帮助](https://support.google.com/gemininotebook/answer/16179559)。

### 音频概览能做什么？

音频概览可以把选中的资料整理成方便听的内容，适合通勤时回顾研究报告或课程。可以先生成一版听主要脉络，再回文档查专业名词、数字和原文限制。产品改名及功能变化见 [Google 的公告](https://blog.google/innovation-and-ai/products/gemini-notebook/notebooklm-gemini-notebook/)。'
WHERE `slug` = 'notebooklm-google'
  AND `name` = 'NotebookLM'
  AND `url` = 'https://notebooklm.google.com'
  AND CAST(`desc` AS BINARY) = '对着文档提问'
  AND CAST(`body` AS BINARY) = 'NotebookLM 把上传的文档当成资料来回答，适合读一份长材料。国内网络经常打不开。

它引用的是你给的文件。文件里没有的句子，不要当成已经查过。';

COMMIT;
