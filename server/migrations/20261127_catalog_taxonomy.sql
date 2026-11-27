-- AI 工具按用途整理。先执行 20261127_catalog_content.sql。旧网址及点击数据继续沿用。

BEGIN;
SELECT pg_advisory_xact_lock(82422402);

UPDATE "categories" SET "sort" = 0, "show_on_home" = TRUE WHERE "slug" = 'ai' AND "kind" = 'section';

UPDATE "categories" SET "name" = '通用助手', "sort" = 1 WHERE "slug" = 'chat' AND "parent_slug" = 'ai' AND "name" = '对话';

UPDATE "categories" SET "name" = '视频与短剧', "sort" = 4 WHERE "slug" = 'draw' AND "parent_slug" = 'ai' AND "name" = '创作';

UPDATE "categories" SET "name" = '图像与设计', "sort" = 3 WHERE "slug" = 'ai-ae0faa3203a9' AND "parent_slug" = 'ai' AND "name" = '绘图';

UPDATE "categories" SET "name" = '智能体与流程', "sort" = 8 WHERE "slug" = 'ai-6bf6d3f3d626' AND "parent_slug" = 'ai' AND "name" IN ('对话', 'AI 助手');

INSERT INTO "categories" ("slug", "name", "parent_slug", "kind", "content_kind", "sort", "show_on_home") VALUES
  ('ai-search', '搜索与研究', 'ai', 'tab', 'links', 2, FALSE),
  ('ai-ae0faa3203a9', '图像与设计', 'ai', 'tab', 'links', 3, FALSE),
  ('ai-audio', '音乐与语音', 'ai', 'tab', 'links', 5, FALSE),
  ('ai-code', '编程与应用', 'ai', 'tab', 'links', 6, FALSE),
  ('ai-office', '写作与办公', 'ai', 'tab', 'links', 7, FALSE),
  ('ai-6bf6d3f3d626', '智能体与流程', 'ai', 'tab', 'links', 8, FALSE)
ON CONFLICT DO NOTHING;

UPDATE "category_guides" SET "title" = 'AI 通用助手' WHERE "category_slug" = 'chat' AND "title" = 'AI 对话助手';

UPDATE "category_guides"
SET "title" = 'AI 视频与短剧工具', "intro" = '从一张图做动效、按提示词出镜头，或从剧本开始制作短剧。选工具前先分清任务是做单镜头、多个镜头，还是需要分镜和剪辑的完整流程。', "home_label" = ''
WHERE "category_slug" = 'draw' AND "title" = 'AI 绘画与视频生成' AND "home_label" = '绘图与视频'
  AND "intro" COLLATE "C" = '做插画、商品图或短视频时，先按文生图、图生图或图生视频挑工具。需要交付成品时，生成前先看导出尺寸、水印和商用条款。';

INSERT INTO "category_guides" ("category_slug", "title", "intro", "home_label") VALUES
  ('ai', 'AI 工具导航', '按用途选 AI 工具。日常提问找通用助手，查资料找搜索与研究，做图、视频或配音进相应创作分类，写代码、做文档和搭流程也有独立入口。', 'AI 工具'),
  ('ai-search', 'AI 搜索与研究工具', '查网页信息、比较来源或找论文时，先按资料类型选工具。核对引用时，优先看结果能否直接跳到原文。', ''),
  ('ai-ae0faa3203a9', 'AI 图像生成与设计工具', '生成插画、编辑商品图、制作海报和探索三维形象，都可以从这个分类开始。已有素材时，优先看工具是否支持参考图和局部调整。', ''),
  ('ai-audio', 'AI 音乐、配音与音频工具', '按制作环节找工具。做配乐看歌曲生成，做旁白看文字转语音，需要整理素材时再看音频剪辑和译制。', ''),
  ('ai-code', 'AI 编程助手与应用生成工具', '手头已有代码仓库时，可选能读取项目和编辑文件的编程助手。想先做一个应用雏形，可以从低代码或自然语言生成应用的工具开始。', ''),
  ('ai-office', 'AI 写作、文档与办公工具', '根据要交付的内容选工具。整理资料看文档问答，起草内容看写作，做汇报看演示文稿的编辑和导出能力。', ''),
  ('ai-6bf6d3f3d626', 'AI 智能体与工作流工具', '需要让 AI 按步骤完成任务时，可按触发方式、可连接的数据和是否要人工确认来比较工具。', '')
ON CONFLICT DO NOTHING;

-- 只有原分类和已核对的站点标识均匹配时才移动。
UPDATE "links" AS site
SET category_slug = target.slug
FROM (
SELECT 'metaso' AS "slug", '秘塔' AS "name", 'https://metaso.cn' AS "url", 'chat' AS "old_category", 'ai-search' AS "new_category"
UNION ALL SELECT 'perplexity', 'Perplexity', 'https://www.perplexity.ai', 'chat', 'ai-search'
UNION ALL SELECT 'n', '纳米AI', 'https://www.n.cn/downloads?src=AIBotCode', 'ai-6bf6d3f3d626', 'ai-search'
UNION ALL SELECT 'yige-baidu', '文心一格', 'https://yige.baidu.com', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'leonardo', 'Leonardo', 'https://leonardo.ai', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'tongyi-aliyun-2', '通义万相', 'https://tongyi.aliyun.com/wanxiang', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'tusi', '吐司', 'https://tusi.cn', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'miaoyaai', '妙呀', 'https://www.miaoyaai.cn', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'ihuiwa', '绘蛙', 'https://www.ihuiwa.com', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'seede', 'Seede', 'https://seede.ai', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'd', '堆友', 'https://d.design', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'studio-tripo3d', 'Tripo', 'https://studio.tripo3d.com', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'miaohua-sensetime', '秒画', 'https://miaohua.sensetime.com', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'admuse-qq', '妙思', 'https://admuse.qq.com', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'designkit', '美图设计室', 'https://www.designkit.cn', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'lovart', 'Lovart', 'https://www.lovart.ai/zh', 'draw', 'ai-ae0faa3203a9'
UNION ALL SELECT 'suno', 'Suno', 'https://suno.com', 'draw', 'ai-audio'
UNION ALL SELECT 'moyin', '魔音工坊', 'https://www.moyin.com', 'draw', 'ai-audio'
UNION ALL SELECT 'haimian', '海绵音乐', 'https://www.haimian.com', 'draw', 'ai-audio'
UNION ALL SELECT 'tianyin-music-163', '网易天音', 'https://tianyin.music.163.com', 'draw', 'ai-audio'
UNION ALL SELECT 'y-qq-2', '未音', 'https://y.qq.com/vemus', 'draw', 'ai-audio'
UNION ALL SELECT 'audioeditor-ximalaya', '音剪', 'https://audioeditor.ximalaya.com', 'draw', 'ai-audio'
UNION ALL SELECT 'voicepica-aliyun', '伶鹊', 'https://voicepica.aliyun.com', 'draw', 'ai-audio'
UNION ALL SELECT 'yizhi-iflyrec', '讯飞译制', 'https://yizhi.iflyrec.com', 'draw', 'ai-audio'
UNION ALL SELECT 'codeflying', '码上飞', 'https://www.codeflying.net', 'chat', 'ai-code'
UNION ALL SELECT 'miaoda', '秒哒', 'https://www.miaoda.cn', 'chat', 'ai-code'
UNION ALL SELECT 'iflow', '心流', 'https://iflow.cn', 'chat', 'ai-office'
UNION ALL SELECT 'cp-baidu', '橙篇', 'https://cp.baidu.com', 'chat', 'ai-office'
UNION ALL SELECT 'ima-qq', 'ima', 'https://ima.qq.com', 'chat', 'ai-office'
UNION ALL SELECT 'notebooklm-google', 'NotebookLM', 'https://notebooklm.google.com', 'chat', 'ai-office'
UNION ALL SELECT 'feixianglaoshi', '飞象老师', 'https://www.feixianglaoshi.com', 'chat', 'ai-office'
UNION ALL SELECT 'kukuai', '库库AI', 'https://kukuai.cn', 'chat', 'ai-office'
UNION ALL SELECT 'turbodesk-xfyun', '讯飞绘文', 'https://turbodesk.xfyun.cn', 'chat', 'ai-office'
UNION ALL SELECT 'xiaomengai', '彩云小梦', 'https://www.xiaomengai.com', 'draw', 'ai-office'
UNION ALL SELECT 'aippt', 'AiPPT', 'https://www.aippt.cn', 'draw', 'ai-office'
UNION ALL SELECT 'zhuque', '朱雀', 'https://matrix.tencent.com/ai-detect/', 'draw', 'ai-office'
UNION ALL SELECT 'coze', '扣子', 'https://www.coze.cn', 'chat', 'ai-6bf6d3f3d626'
UNION ALL SELECT 'fellou', 'fellou', 'https://fellou.ai', 'chat', 'ai-6bf6d3f3d626'
UNION ALL SELECT 'tabbit', 'Tabbit', 'https://www.tabbit.com', 'chat', 'ai-6bf6d3f3d626'
) AS change_set
JOIN "categories" AS target ON target.slug = change_set.new_category AND target.parent_slug = 'ai'
WHERE site.slug = change_set.slug AND site.name COLLATE "C" = change_set.name
  AND site.url COLLATE "C" = change_set.url AND site.category_slug = change_set.old_category;

INSERT INTO "links" ("slug", "category_slug", "name", "url", "desc", "tags", "alias", "sort", "body",
  "status", "clicks", "views", "check_fails", "user_id", "pinned", "pin_sort", "created_at")
SELECT seed.slug, seed.category_slug, seed.name, seed.url, seed."desc", seed.tags, seed.alias, seed.sort, seed.body,
  'online', 0, 0, 0, 0, FALSE, 0, CURRENT_TIMESTAMP(3)
FROM (
SELECT 'cursor-ai' AS "slug", 'ai-code' AS "category_slug", 'Cursor' AS "name", 'https://cursor.com/' AS "url", 'AI 代码编辑器，支持理解项目、修改文件和审阅变更' AS "desc", '编程,代码编辑器,AI' AS "tags", '' AS "alias", 1 AS "sort", 'Cursor 适合在已有代码仓库里理解项目、规划改动和审阅代码差异。它提供能编辑文件的 Agent，也提供可先讨论方案的模式。

### 第一次使用怎么控制改动范围？

打开一个熟悉的仓库，先让它找出某个页面的入口和测试。确认位置后，指定一个具体的小改动，例如补一个空状态，然后查看逐文件差异并运行相关测试。涉及多个文件时，可先让它列方案再选择范围。[Cursor 官方 Agent 指南](https://cursor.com/help/ai-features/agent)' AS "body"
UNION ALL SELECT 'github-copilot', 'ai-code', 'GitHub Copilot', 'https://github.com/features/copilot', 'GitHub 的编程助手，支持 IDE 代码补全与代码问答', '编程,代码补全,AI', '', 2, 'GitHub Copilot 可以在支持的 IDE 里随输入给出行内代码建议，也能用对话解释已有代码或讨论实现。它适合希望留在现有编辑器里完成日常编程的人。

### 代码补全和对话分别什么时候用？

写一个小函数时，可以先写函数名、参数和预期返回值，再看行内建议。遇到陌生模块时，把相关代码带进对话，让它解释数据从哪里来、应在哪个位置补测试。采纳建议前看差异并跑测试。[GitHub IDE 补全指南](https://docs.github.com/en/copilot/how-tos/copilot-in-your-ide/copilot-for-common-tasks/get-ide-code-suggestions) · [IDE 对话指南](https://docs.github.com/en/copilot/how-tos/copilot-in-your-ide/chat-with-copilot/chat-in-ide)'
UNION ALL SELECT 'elicit', 'ai-search', 'Elicit', 'https://elicit.com/', '用研究问题找论文，支持文献筛选和证据提取', '论文,文献综述,AI', '', 3, 'Elicit 的对象是研究论文。它可以根据研究问题查找文献，也提供为系统综述设计的筛选、提取和报告流程。

### 怎么用它整理一组论文？

先把问题写得能用文献回答，例如限定对象、干预和结果。筛出相关论文后，可以按样本、方法、结论等字段做提取表。发现结果冲突时，回论文的摘要、方法和数据表核对，尤其留意工具是否拿到了全文。[Elicit 系统综述手册](https://support.elicit.com/en/articles/14759154-systematic-reviews-in-elicit)'
UNION ALL SELECT 'gamma-app', 'ai-office', 'Gamma', 'https://gamma.app/', '把提纲或已有内容整理成演示文稿，支持导出 PPTX', 'PPT,演示文稿,AI', '', 4, 'Gamma 可以从提示词生成演示文稿，也可以导入已有文档或幻灯片作为起点。适合准备汇报初稿、课程讲义和介绍材料。

### 如何把提纲变成一份可交付的 PPT？

先写受众、目的和各页要讲的结论，例如用五页解释一个项目的背景、进度和下一步。生成后检查标题层级和图表说明，再导出 PPTX 或 PDF。接收方需要二次编辑时，优先试导出 PPTX，并在目标软件里检查一次排版。[Gamma 创建指南](https://help.gamma.app/en/articles/7838093-how-do-i-create-a-new-presentation-document-or-webpage-in-gamma) · [导出帮助](https://help.gamma.app/en/articles/8022861-what-s-the-easiest-way-to-export-my-gamma)'
UNION ALL SELECT 'elevenlabs', 'ai-audio', 'ElevenLabs', 'https://elevenlabs.io/', 'AI 文字转语音，适合旁白、配音和分段试听', '配音,文字转语音,AI', '', 5, 'ElevenLabs 提供文字转语音和浏览器音频制作工具。可以把脚本转成旁白，再试听语速、停顿和专名的读法。

### 做旁白时怎么试音？

先用一小段实际台词选音色，并把人名、产品名或容易念错的缩写放进样本。试听满意后按段生成，方便单独重做有问题的部分。多人对话可先按角色把脚本分好，再决定各段音色。[ElevenLabs 文字转语音文档](https://elevenlabs.io/docs/overview/capabilities/text-to-speech)'
UNION ALL SELECT 'adobe-firefly', 'ai-ae0faa3203a9', 'Adobe Firefly', 'https://firefly.adobe.com/', 'Adobe 的 AI 图像工具，支持文字出图与选区编辑', '图像生成,局部编辑,AI', '', 6, 'Adobe Firefly 有文字出图和生成式编辑入口。做一张新图时可以从构图、主体和光线写提示词。手头已有图片时，可以选择要改的区域，再指定要添加或替换的内容。

### 一张产品图只想换背景时怎么做？

优先用已有图做选区编辑。只圈背景，保留商品轮廓，描述新背景的颜色和光线。生成之后放大检查边缘和阴影，再把合适的结果用于下一版排版。[Adobe Firefly 产品页](https://www.adobe.com/products/firefly.html)'
UNION ALL SELECT 'dify', 'ai-6bf6d3f3d626', 'Dify', 'https://www.dify.ai/', '用可视化流程搭建 AI 应用，支持接入知识库', '工作流,知识库,智能体,AI', '', 7, 'Dify 可以在画布上连起模型、知识检索和分支等步骤。它适合需要把问答或文档处理整理成可重复流程的人。

### 知识库问答从哪里开始？

用几份允许使用的文档建一个小知识库，先接入检索再生成回答。测试两类问题，一类能从文档找到答案，另一类文档没有涉及。查看检索到了什么，以及模型遇到缺信息时是否会说明。[Dify 入门说明](https://www.dify.ai/blog/get-started-with-dify) · [工作流说明](https://www.dify.ai/workflows)'
UNION ALL SELECT 'n8n', 'ai-6bf6d3f3d626', 'n8n', 'https://n8n.io/', '节点式自动化工作流，适合连接应用和定时任务', '自动化,工作流,AI', '', 8, 'n8n 用触发器和节点组织自动化流程。它适合连接已有应用、按时处理数据，再把结果保存或通知给使用者。

### 第一条工作流可以做什么？

先做一条只有几步的流程，例如定时读取公开天气数据，提取当天的摘要，再写入数据表。先手动跑一遍，逐节点检查输入输出，确认格式和失败情况后再开启定时运行。[n8n 官方工作流入门](https://github.com/n8n-io/n8n-docs/blob/main/docs/get-started/build-your-first-workflow.md)'
) AS seed
JOIN "categories" AS category ON category.slug = seed.category_slug AND category.parent_slug = 'ai'
WHERE NOT EXISTS (
  SELECT 1 FROM "links" AS existing
  WHERE existing.slug = seed.slug OR lower(existing.name) = lower(seed.name)
    OR TRIM(TRAILING '/' FROM existing.url) = TRIM(TRAILING '/' FROM seed.url)
);

COMMIT;
