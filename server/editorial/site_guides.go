package editorial

import (
	"fmt"

	"opentools/models"
	"opentools/seed"

	"gorm.io/gorm"
)

type siteGuide struct {
	slug, name, url, oldDesc, oldBody, desc, body string
}

// BackfillSiteGuides updates only the original catalog text for known entries.
func BackfillSiteGuides(db *gorm.DB) error {
	for _, guide := range siteGuides {
		if err := db.Model(&models.Link{}).
			Where("slug = ? AND name = ? AND url = ?", guide.slug, guide.name, guide.url).
			Where("CAST(`desc` AS BINARY) = ?", guide.oldDesc).
			Where("(body IS NULL OR body = '' OR CAST(body AS BINARY) = ?)", guide.oldBody).
			Updates(map[string]any{"desc": guide.desc, "body": guide.body}).Error; err != nil {
			return fmt.Errorf("update site guide for %s: %w", guide.name, err)
		}
	}
	return nil
}

var siteGuides = []siteGuide{
	{
		slug: "poe", name: "Poe", url: "https://poe.com", oldDesc: "多个模型对话",
		oldBody: seed.LegacyPlaceholderIntro("Poe", "多个模型对话"),
		desc:    "多模型 AI 聊天平台，可选 GPT、Claude 和 Gemini",
		body: `Poe 是 Quora 提供的多模型 AI 聊天平台，同一个账号可以体验 GPT、Claude、Gemini 等模型，也可以使用官方或用户创建的机器人。它适合用同一份任务对比几个模型的回答，或试用专门做图片、视频和写作的机器人。

### Poe 怎么选模型？

打开 [Poe 机器人目录](https://poe.com/explore)，搜索模型名，点进机器人就能看它适合做什么。想比较 GPT 和 Claude 的中文改写效果，可以分别新建对话，把同一段文字和同样的要求发过去，再对照事实有没有被改动、语气是否自然。新开对话可以避免之前的聊天内容影响比较。

### Poe 免费吗？积分怎么扣？

免费账号会收到每日积分，未使用的每日积分不会累积。各机器人扣分规则不同，有的按一次生成扣分，有的会随输入和输出长度变化。做长文总结或生成视频前，先点机器人的名称看具体规则，再决定用哪个模型。订阅提供更多积分和部分模型的使用权限。具体额度可查 [Poe 官方积分与机器人说明](https://help.poe.com/hc/en-us/articles/19944206309524-Poe-FAQs)。

### 怎么创建自己的机器人？

从 [创建机器人](https://poe.com/create_bot) 选择基础模型，写下它该怎么回答。例如，建一个中英文邮件修改助手，要求每次先列出容易误解的句子，再给出正式和简短两版。试用阶段可以先设为私有。更多创建方式见 [Poe 创作者文档](https://creator.poe.com/docs)。`,
	},
	{
		slug: "gemini-google", name: "Gemini", url: "https://gemini.google.com", oldDesc: "对话",
		oldBody: seed.LegacyPlaceholderIntro("Gemini", "对话"),
		desc:    "Google AI 助手，支持对话与文件分析",
		body: `Gemini 是 Google 的 AI 助手，可在网页上起草文字、改写、分析文件。比如把两份公开的 PDF 放在同一轮对话里，先让它分别概括，再按关注的几个问题比较差异。

### Gemini 怎么读 PDF 和表格？

在对话框旁添加文件，再说明要查什么。读报告时，可以要求按「结论、原文依据、页码」输出。看表格时，先请它写清用了哪些列、单位和筛选条件，再把关键数字回到原表复核。支持的文件和上传限额见 [Gemini 文件分析帮助](https://support.google.com/gemini/answer/14903178)。

### Gemini 和 NotebookLM 怎么选？

临时改一段文字或问一份附件，可以先用 Gemini 对话。若要围绕一组论文或课件反复提问、核对出处，适合把这些资料放进 [Gemini Notebook（原 NotebookLM）](/site/notebooklm-google)，日后继续在同一个笔记本里查。`,
	},
	{
		slug: "copilot-microsoft", name: "Copilot", url: "https://copilot.microsoft.com", oldDesc: "对话",
		oldBody: seed.LegacyPlaceholderIntro("Copilot", "对话"),
		desc:    "微软个人版 AI 助手，支持网页问答与文件分析",
		body: `Microsoft Copilot 的个人网页入口是 copilot.microsoft.com，可以用来提问、起草内容，也可以上传支持的文件或图片让它分析。短任务可以直接在网页上开始，登录个人 Microsoft 账号后还能使用聊天记录等功能。

### 网页版 Copilot 可以怎么用？

拿一份公开的活动介绍试试，让它按「时间、地点、报名截止、费用」提取，缺失的项目标为未说明。看网页总结时，可以在 Edge 中用 Copilot 整理当前页面，再回原文核对数字和关键条款。功能入口和文件支持见 [Microsoft Copilot 功能说明](https://www.microsoft.com/en/microsoft-copilot/for-individuals/features)。

### Copilot 和 Microsoft 365 Copilot 有什么区别？

个人网页版适合日常问答和写作。想在 Word 或 Excel 里直接用 Copilot，要看 Microsoft 365 订阅及账号权限。工作或学校账号由组织管理，能否使用及能访问哪些资料取决于组织配置。可参照 [微软的版本说明](https://support.microsoft.com/en-us/microsoft-365-copilot/what-s-the-difference-between-microsoft-copilot-free-and-copilot-in-microsoft-365)。`,
	},
	{
		slug: "notebooklm-google", name: "NotebookLM", url: "https://notebooklm.google.com", oldDesc: "对着文档提问",
		oldBody: "NotebookLM 把上传的文档当成资料来回答，适合读一份长材料。国内网络经常打不开。\n\n它引用的是你给的文件。文件里没有的句子，不要当成已经查过。",
		desc:    "Gemini Notebook，按来源做文档问答与音频概览",
		body: `NotebookLM 已更名为 Gemini Notebook，是 Google 的资料研究工具。把 PDF、网页或支持的 Google Drive 文件放进同一个笔记本，就可以围绕这组资料问问题、查看回答引用的原文，还可以生成音频概览和复习材料。

### NotebookLM 怎么用来读论文？

在 [Gemini Notebook](https://notebook.google.com/) 新建笔记本，把同一主题的几篇论文分别添加为来源。可以问「请按论文列出样本量、研究对象和结论，并给出原文引用」。随后点引用回到对应资料。只想核对其中一篇时，在来源面板只选那篇。操作可查 [Google 的来源管理说明](https://support.google.com/gemininotebook/answer/16215270) 和 [对话与引用帮助](https://support.google.com/gemininotebook/answer/16179559)。

### 音频概览能做什么？

音频概览可以把选中的资料整理成方便听的内容，适合通勤时回顾研究报告或课程。可以先生成一版听主要脉络，再回文档查专业名词、数字和原文限制。产品改名及功能变化见 [Google 的公告](https://blog.google/innovation-and-ai/products/gemini-notebook/notebooklm-gemini-notebook/)。`,
	},
}
