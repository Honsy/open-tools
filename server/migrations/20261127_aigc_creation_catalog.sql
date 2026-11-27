-- 补充创作工具，并替换已核对的占位介绍。保留既有网址、上架状态和访问数据。

BEGIN;
SELECT pg_advisory_xact_lock(82422402);

UPDATE "links" AS site
SET "desc" = guide."desc", tags = guide.tags, body = guide.body
FROM (
  SELECT 'capcut' AS "slug", 'draw' AS "category_slug", '剪映' AS "name", 'https://www.capcut.cn' AS "url", '视频剪辑' AS "old_desc", '视频' AS "old_tags", '给口播识别字幕，结合多轨时间线编辑短片和生成素材' AS "desc", '视频剪辑,智能字幕,多轨剪辑,AIGC' AS "tags", '剪映专业版可把拍摄素材、生成镜头、旁白和配乐放在多轨时间线上继续剪辑。它提供智能字幕，用于识别音轨里的说话内容。适合把零散镜头整理成可以发布的短视频。

### 已生成几条镜头，怎么拼成一段口播？

先搭主线并把旁白放到需要出现的位置，再让剪映识别字幕。校对人名和品牌，把过长的字幕拆到便于阅读的节奏。用画面覆盖停顿或跳剪处，横屏改竖屏时再检查标题和字幕是否处于安全区。导出一版在手机上完整回放，有助于找出配乐过响和字幕遮挡的问题。[剪映专业版功能说明](https://www.jianying.com/web/professional)' AS "body"
  UNION ALL SELECT 'hailuoai', 'draw', '海螺AI', 'https://hailuoai.com', '视频', '视频,AI', 'MiniMax 的视频创作工具，适合做图片动效、运镜和首尾帧过渡', '图生视频,首尾帧,短视频,AIGC', '海螺 AI 是 MiniMax 的创作入口，可以从文字或图片开始做视频镜头。官网的创作示例有首尾帧、风格转绘和广告电商等用法。手里已有插画或产品图时，可以先做一个动作明确的小样。

### 有一张作品图，怎么让它动起来？

从「瓶身缓慢转半圈，镜头固定」这类单一动作开始，方便检查商品形状和文字是否变形。需要把镜头落在一张既定画面时，可以在支持首尾帧的模式里放入两张构图相近的图，并写清中间的变化。拿短片做广告前逐帧检查品牌、人物轮廓和关键道具。[海螺 AI 官网与创作示例](https://hailuoai.com/)'
  UNION ALL SELECT 'runwayml', 'draw', 'Runway', 'https://runwayml.com', '视频', '视频,AI', '从文字或参考图生成镜头，适合单独测试主体动作和运镜', '图生视频,文生视频,运镜,AIGC', 'Runway 的创作工作台支持图生视频和文生视频。用参考图开场时，构图、主体和光线主要由输入图提供，提示词可以集中描述主体怎么动、机位怎么变，以及动作何时结束。

### 怎样写一条能看出效果的运镜提示？

拿一张边缘清晰的街景图，先试「镜头沿人行道慢慢向前，远处行人走过」。如果需要静止的机位，只改镜头这部分，再和第一版比较。在编辑器里相邻摆放两条镜头，检查运动方向能否衔接，以及背景变化会不会抢走人物动作。[Runway 图生视频指南](https://help.runwayml.com/hc/en-us/articles/48324313115155-Image-to-Video-Prompting-Guide)'
  UNION ALL SELECT 'leonardo', 'ai-ae0faa3203a9', 'Leonardo', 'https://leonardo.ai', '图像', '图像,AI', '参考图生成插画与概念图，支持标注区域后继续修改', 'AI绘图,概念设计,参考图,局部编辑,AIGC', 'Leonardo.Ai 适合做插画、概念设计和风格探索。使用支持参考图和 Annotate 的模型时，可以把已满意的一版作为输入，标出下一步需要调整的区域，再描述要保留的构图或物件。

### 角色已经画好了，只改手里的道具怎么办？

先保留角色原图，再尝试只在道具位置做标注，用短句描述新道具的形状和材质。对照原图检查手指、袖口和背景边缘，看看改动有没有影响到其他区域。要做多轮变体时，把已通过的一版保存下来，方便比较同一角色在不同场景里的效果。[Leonardo 图像编辑示例](https://leonardo.ai/gpt-image-2-6)'
  UNION ALL SELECT 'tongyi-aliyun-2', 'ai-ae0faa3203a9', '通义万相', 'https://tongyi.aliyun.com/wanxiang', '图像和视频', '图像,AI', '阿里的图像和视频创作入口，适合按文字出图并用参考帧做动态素材', '文生图,图像编辑,图生视频,首尾帧,AIGC', '通义万相是阿里的图像与视频创作入口。可以先通过文字描述探索插画或商品场景，再把选中的画面用于动态素材。阿里云文档还列有图像编辑、首帧和首尾帧生视频等模型能力。

### 怎么把一张场景图做成短视频素材？

先挑出构图稳定的一张图，确认主体、画幅和要保留的细节。只有一张图时，试写一个主要动作。如果手头有两张构图接近的起止图，可以在支持首尾帧的模型中尝试补出过渡。不同模型支持的输入不同，使用网页或 API 前先确认所选模式。[万相创作官网](https://tongyi.aliyun.com/wan) · [阿里云视频提示词指南](https://help.aliyun.com/zh/model-studio/wan3-video-generation-prompt-guide)'
  UNION ALL SELECT 'liblib-3', 'ai-ae0faa3203a9', 'LibLib', 'https://www.liblib.art', 'AI绘画原创模型分享社区，10万+模型免费下载', 'AIGC', '中文 AI 绘画模型社区，支持在线生图和 ComfyUI 工作流', 'AI绘图,模型社区,ComfyUI,节点工作流,AIGC', 'LiblibAI（哩布哩布 AI）把模型社区和在线生成放在同一个网站。可以先看模型的示例图和使用条件，再用 WebUI 做第一张图。想复用多个处理步骤时，可以打开在线 ComfyUI 工作流。

### 同一个提示词，怎样比较两种模型风格？

挑同一类型的两个模型，用相同主体描述和画幅各试一张，看看用色、构图和人物轮廓的区别。找到合适的模型后，再记下生成参数。要复用社区工作流时，可以先看它需要上传哪些素材、会输出哪些文件，并检查模型的使用条件。[LiblibAI 官方在线生图教程](https://www.liblib.art/tutorial/1)'
) AS guide
WHERE site.slug = guide.slug AND site.category_slug = guide.category_slug
  AND site.name COLLATE "C" = guide.name
  AND TRIM(TRAILING '/' FROM split_part(site.url, '?', 1)) COLLATE "C" = guide.url
  AND site."desc" COLLATE "C" = guide.old_desc AND site.tags COLLATE "C" = guide.old_tags
  AND site.status = 'online'
  AND (
    COALESCE(site.body, '') = ''
    OR site.body COLLATE "C" = CONCAT(
      guide.name, '，', guide.old_desc, '。先看它是不是你要找的那一个，再打开。',
      chr(10), chr(10), '页面改版、要登录或分地区开放时，以网站自己的说明为准。'
    )
    OR (guide.slug = 'liblib-3' AND site.body COLLATE "C" = guide.old_desc)
  );

INSERT INTO "links" ("slug", "category_slug", "name", "url", "desc", "tags", "alias", "lang", "region", "sort", "body",
  "status", "clicks", "views", "check_fails", "user_id", "pinned", "pin_sort", "created_at")
SELECT seed.slug, seed.category_slug, seed.name, seed.url, seed."desc", seed.tags, seed.alias, seed.lang, seed.region, seed.sort, seed.body,
  'online', 0, 0, 0, 0, FALSE, 0, CURRENT_TIMESTAMP(3)
FROM (
  SELECT 'comfyui' AS "slug", 'ai-ae0faa3203a9' AS "category_slug", 'ComfyUI' AS "name", 'https://comfy.org/' AS "url", '用节点搭建可复用的生图与素材处理流程，支持本地或云端运行' AS "desc", '节点工作流,AI绘图,本地生图,AIGC' AS "tags", 'Comfy' AS "alias", '' AS "lang", '' AS "region", 7 AS "sort", 'ComfyUI 是用节点组织生成和处理步骤的创作工具。模型、提示词、采样和输出各有对应节点。可以用官方模板起步，把跑通的流程保存下来，下次只换输入或参数，适合反复制作同类素材。

### 第一套工作流怎么搭？

从官方文生图教程开始，加载一个可用模型，用短提示词生成并保存一张图片。记下模型和尺寸，把工作流保存为 JSON。下次做同系列封面时，重新打开工作流，只替换主体描述，比较不同素材的构图和留白。如果导入别人的流程后提示缺少节点，先看对应模板的依赖说明。[ComfyUI 首张图教程](https://docs.comfy.org/get_started/first_generation)' AS "body"
  UNION ALL SELECT 'krea', 'ai-ae0faa3203a9', 'Krea', 'https://www.krea.ai/', '根据草图实时预览带风格的画面，也可继续放大和增强', '实时生图,草图转图,图片放大,AIGC', '', '', '', 8, 'Krea 的 Realtime 功能会随草图变化更新预览，适合在定稿前试不同的主体位置、形状和色块。看中某个方向后，可以再做图像编辑或放大，补足成品所需的细节。

### 怎么从几笔草图挑一个封面构图？

在画布上用大色块标出主体、背景和标题留白，再用提示词写清材质和光线。每次只移动一个元素，边改边看预览。从缩略图里挑一版主次清楚的构图，再放大检查手部、边缘或小纹理。标题可在排版阶段单独加上，方便后续改字。[Krea 官方功能介绍](https://www.krea.ai/) · [Realtime 入口](https://www.krea.ai/realtime)'
  UNION ALL SELECT 'photoroom', 'ai-ae0faa3203a9', 'Photoroom', 'https://www.photoroom.com/', '商品图抠图、生成场景背景和批量统一版式', '商品图,电商,抠图,换背景,AIGC', '', '', '', 9, 'Photoroom 面向电商图和目录素材。可以把拍好的商品从背景中分离，再生成新的场景背景。需要制作一批同类商品图时，可以用批处理统一背景和画布摆放。

### 怎样给一组商品图换统一背景？

先选一张边缘清晰的实拍作为样例，生成浅色背景，并核对商标、颜色、包装说明和阴影。确定留白和构图后，再把相同版式用于整组商品。透明瓶身、反光部件和薄边缘值得单独放大检查。需要对照平台主图规范时，留一份原始照片和白底版。[Photoroom AI 背景说明](https://www.photoroom.com/tools/instant-backgrounds)'
  UNION ALL SELECT 'meshy', 'ai-ae0faa3203a9', 'Meshy', 'https://www.meshy.ai/', '把单张或多角度照片转成 3D 模型，可导出 GLB、FBX 等格式', '图片转3D,三维建模,GLB,游戏素材,AIGC', '', '', '', 12, 'Meshy 可以从文字或参考图片生成带纹理的 3D 模型。Image to 3D 支持多视图输入，适合先做游戏道具、网页展示物件或产品外形样稿，再在三维软件里继续修整。

### 一张正面图够不够？

如果只需要快速看轮廓，可以先用一张主体居中的照片。需要还原背面或侧边时，再补同一物体的背面和侧面照片，尽量统一光线。生成后旋转模型检查背面、孔洞和纹理接缝。网页展示可试 GLB，进游戏引擎前再核对面数和材质。[Meshy 图片转 3D 指南](https://www.meshy.ai/features/image-to-3d)'
  UNION ALL SELECT 'luma-creative', 'draw', 'Luma', 'https://lumalabs.ai/app', '用既有视频改画风，也能补全画面适配横竖屏', '视频风格化,视频扩展,视频生成,AIGC', 'Dream Machine', '', '', 10, 'Luma 的网页创作工作台支持文生视频、图生视频，也提供修改既有镜头的 Modify Video 和扩展画幅的 Reframe。有真拍或预演素材时，可以沿用原来的动作节奏，试做另一种场景或风格。

### 怎样把实拍片段改成一种画风？

选一条运镜稳定、主体清楚的短素材。先试一种画风和一个可识别的背景变化，再对照源片检查人物动作、手部、道具和镜头时序。要把横屏片段改成竖屏时，可以另试 Reframe，在手机尺寸检查主体位置和扩出来的区域是否接得上。[Luma 视频能力说明](https://lumalabs.ai/learning-hub/luma-video-capabilities-ai-video-generation-editing-lip-sync)'
  UNION ALL SELECT 'ltx-studio', 'draw', 'LTX Studio', 'https://ltx.io/studio', '从剧本做视频分镜，复用角色素材并在时间线上预览', '剧本分镜,角色一致性,短剧,短片,AIGC', '', '', '', 11, 'LTX Studio 把分镜、素材和时间线放在一个视频制作工作台里。它的 Elements 功能可以把角色保存成素材，供不同镜头引用。适合先看短片的叙事和节奏，再挑关键镜头做视频生成。

### 怎么把一个故事做成可讨论的分镜？

写出每场发生的动作，把一场拆成建立场景、角色行动和反应三个镜头。给反复出镜的人物建一个 Element，在分镜里引用并核对服装和道具。把镜头在时间线上按顺序播放，确认没有缺少关键信息，再生成成本较高的段落。[LTX Studio 官方教程](https://ltx.io/blog/ltx-studio-tutorial) · [帮助中心](https://help.ltx.io/hc/en-us/p/ltx-studio)'
  UNION ALL SELECT 'pixverse', 'draw', 'PixVerse', 'https://pixverse.ai/en', '文字或图片生成视频，可用支持的模型在首尾帧之间做转场', '首尾帧,转场,图生视频,AIGC', '', '', '', 12, 'PixVerse 的创作入口包括文生视频、图生视频和视频模板。已有开头和结尾两张画面时，可以选择支持首尾帧的模型补中间动作，适合短转场或连续镜头测试。

### 怎样用两张图做一段可剪进去的转场？

先选画幅一致、共享清晰视觉元素的两张图，例如窗外白天和夜晚各一张。描述该怎样变化，比如光线慢慢暗下、镜头保持固定。重点检查中间帧里的窗框和道具是否扭曲，再把短片接到前后镜头试听或预览。[PixVerse 创作入口](https://pixverse.ai/en) · [模型能力对照](https://docs.platform.pixverse.ai/model-capability-map-2394982m0)'
  UNION ALL SELECT 'descript', 'draw', 'Descript', 'https://www.descript.com/', '把访谈或口播转成文字，通过编辑转写稿同步剪视频', '文字剪辑,口播剪辑,字幕,播客,AIGC', '', '', '', 13, 'Descript 会把导入的音视频转成文字稿。删掉稿子里的某段话时，可以同步删掉对应片段。适合访谈、录屏教学和以讲述为主的视频，也提供字幕和口头语处理入口。

### 怎么把一段访谈剪成短视频？

先在转写稿里标出一个能独立理解的观点，保留必要的提问或背景。清理重复句后回放切点，检查是否截断呼吸或带走了指代对象。最后核对字幕里的人名和术语。如果剪掉一段后画面跳动明显，可以用相关画面覆盖剪切处。[Descript 视频编辑说明](https://www.descript.com/video-editing)'
  UNION ALL SELECT 'mureka', 'ai-audio', 'Mureka', 'https://www.mureka.ai/', '用自定义歌词、音乐风格或纯音乐模式生成歌曲和配乐', 'AI作曲,歌词生成歌曲,配乐,AIGC', '', '', '', 8, 'Mureka 的自定义模式可以分别填写歌词和音乐风格，也可切换成纯音乐。可以从模板起步，再调乐器、节奏和情绪，适合试作歌曲小样和不同情绪的配乐。

### 有一段歌词，怎样试听两种编曲？

把主歌和副歌分段写清楚，用相同歌词分别试轻声原声编曲和节拍强的编曲。试听时重点看重音是否落在正确的词上、副歌是否容易辨认。如果是给有旁白的视频用，可以再生成一版纯音乐，并在视频里检查它是否盖住人声。[Mureka 模板与自定义指南](https://www.mureka.ai/blog_en/how-to-use-mureka-templates.html)'
  UNION ALL SELECT 'aiva', 'ai-audio', 'AIVA', 'https://www.aiva.ai/', '按风格生成音乐，可用 MIDI 或音频作为作曲参考，再编辑轨道', 'AI作曲,MIDI,编曲,配乐,AIGC', '', '', '', 9, 'AIVA 适合需要继续修改编曲的创作者。可以从预设风格开始，也可以上传自己有权使用的音频或 MIDI 作为参考。生成后可编辑轨道，按实际交付需要选择导出格式。

### 已经有个旋律，接下来怎么做？

用一段自写的短旋律做 MIDI 参考，先试和画面情绪相近的配器。试听时标记需要保留的动机，并听低音是否遮住旋律。要送进制作软件时，确认账户的导出格式和作品许可，再把音轨或 MIDI 导入目标项目检查。[AIVA 官方产品与方案](https://www.aiva.ai/)'
  UNION ALL SELECT 'soundraw', 'ai-audio', 'SOUNDRAW', 'https://soundraw.io/', '为短片生成背景音乐，可调时长、段落强度和乐器', '背景音乐,短片配乐,混音,AIGC', '', '', '', 10, 'SOUNDRAW 面向需要为视频、播客或广告找配乐的创作者。可以按曲风生成音乐，再在 Mixer 调整时长、乐器和段落强度，适合已有粗剪、想让音乐贴合画面节奏的场景。

### 怎样配一条有人声的产品视频？

把时间线分成开场、讲解、结尾，先按视频时长生成一版。讲解段降低配乐强度，结尾再增加鼓点。试听人声与音乐叠在一起的效果，并把停顿处和最后一帧一起检查。用于客户项目或付费推广前，按用途查对应的许可范围。[SOUNDRAW 调整配乐说明](https://soundraw.io/)'
  UNION ALL SELECT 'stable-audio', 'ai-audio', 'Stable Audio', 'https://stableaudio.com/', '根据文字生成拟音、环境声、转场音效和音乐素材', '音效生成,环境声,拟音,AIGC', '', '', '', 11, 'Stable Audio 可以用文字生成音乐和音效。比如画面里已有物件动作但没录到声音时，可以做一组短拟音，再把合适的一条放回剪辑时间线。

### 怎样补一个柜门关上的声音？

提示词写清发声的物体、动作和空间，比如「木柜门快速合上，闷响，室内短混响」。把柜门声与持续的室内底噪分开生成，剪辑时分别对齐动作点和整段画面。试听时听尾声是否自然，也可以尝试同一描述加上「近距离」来比较听感。[Stability AI 音效提示词指南](https://stability.ai/guides/stable-audio-3-prompt-guide)'
  UNION ALL SELECT 'liblib-3', 'ai-ae0faa3203a9', 'LibLib', 'https://www.liblib.art/', '中文 AI 绘画模型社区，支持在线生图和 ComfyUI 工作流', 'AI绘图,模型社区,ComfyUI,节点工作流,AIGC', 'LiblibAI', '中文', '中国', 0, 'LiblibAI（哩布哩布 AI）把模型社区和在线生成放在同一个网站。可以先看模型的示例图和使用条件，再用 WebUI 做第一张图。想复用多个处理步骤时，可以打开在线 ComfyUI 工作流。

### 同一个提示词，怎样比较两种模型风格？

挑同一类型的两个模型，用相同主体描述和画幅各试一张，看看用色、构图和人物轮廓的区别。找到合适的模型后，再记下生成参数。要复用社区工作流时，可以先看它需要上传哪些素材、会输出哪些文件，并检查模型的使用条件。[LiblibAI 官方在线生图教程](https://www.liblib.art/tutorial/1)'
) AS seed
JOIN "categories" AS category ON category.slug = seed.category_slug AND category.parent_slug = 'ai'
WHERE NOT EXISTS (
  SELECT 1 FROM "links" AS existing
  WHERE existing.slug = seed.slug OR existing.name = seed.name
    OR existing.alias = seed.name OR (seed.alias <> '' AND existing.name = seed.alias)
    OR TRIM(TRAILING '/' FROM split_part(existing.url, '?', 1)) = TRIM(TRAILING '/' FROM seed.url)
);

COMMIT;
