let snapshot = null;
let currentView = "todo";
let cy = null;
let selectedNodeID = null;
let activeInspectorTab = "info";
let activeTruthOwnerID = null;
let activeDocMode = "rendered";
let mermaidReady = false;
let activeSpecflowNavGroup = "unit";
let activeReviewNavGroup = "candidate";
let activeTodoNavGroup = "stableVerify";
let snapshotRequestInFlight = false;
let snapshotDataSignature = "";
let activeSourceHeadings = [];
let docGuideOpen = false;

const LANGUAGE_STORAGE_KEY = "specflow-reader-language";
const SUPPORTED_LANGUAGES = ["zh-CN", "en"];
const SNAPSHOT_POLL_INTERVAL_MS = 5000;
let currentLanguage = readStoredLanguage();

const TRANSLATIONS = {
  "zh-CN": {
    loading: "读取中",
    viewNavAria: "Reader 视图",
    graphLegendAria: "图节点颜色说明",
    resizeAria: "调整检查面板宽度",
    inspectorTabsAria: "节点检查",
    docModeAria: "Spec 文档显示模式",
    docGuideAria: "Spec 文档导览",
    refresh: "刷新",
    language: {
      label: "语言",
      zh: "中文"
    },
    tabs: {
      todo: { title: "待处理", subtitle: "下一步" },
      spec: { title: "Spec 查看", subtitle: "文档入口" },
      status: { title: "状态", subtitle: "进度对齐" },
      project: { title: "项目结构", subtitle: "仓库路径" },
      specflow: { subtitle: "治理层级" }
    },
    legend: {
      unit: {
        label: "单元",
        tooltip: "单元是一块可独立说明、开发和验证的工程责任，例如 agent、memory 或 tool。"
      },
      scenario: {
        label: "场景",
        tooltip: "场景是一条从触发到结果的完整使用链路，用来说明多个责任块怎样一起完成一个结果。"
      },
      rule: {
        label: "规则",
        tooltip: "规则是多个单元或场景共同复用的一段规则，避免同一规则在不同地方重复写。"
      },
      truth: {
        label: "Spec 文档",
        tooltip: "Spec 文档是当前项目认可的说明文档；开发和验证都要以这些文档为准。"
      },
      path: {
        label: "实现路径",
        tooltip: "实现路径是代码、配置或工具文件所在的位置，表示规则落到哪些工程文件里。"
      },
      system: {
        label: "全局规则",
        tooltip: "全局规则是全仓库通用的技术底线，例如默认选择、禁止事项和全局例外。"
      }
    },
    views: {
      todo: {
        title: "待处理",
        summary: "从状态索引汇总每个对象的下一步动作。这里是统一入口，先看要做什么，再打开对应材料。",
        nav: "下一步动作"
      },
      spec: {
        title: "Spec 查看",
        summary: "查看当前需要确认的 candidate Spec，以及已经成为 stable 的正式 Spec。完整下一步动作请看“待处理”。",
        nav: "Spec 文档"
      },
      project: {
        title: "项目结构",
        summary: "从仓库路径看实现位置：哪些代码或工程路径已经归到具体责任对象，先不展示 SpecFlow 自己的 Spec 文档和支撑文件。",
        nav: "目录",
        groups: {
          areas: "实现区域"
        }
      },
      specflow: {
        title: "SpecFlow",
        summary: "从治理层级看规则：全局规则、项目映射、状态索引、规则、单元和 Spec 文档如何分层。",
        nav: "对象"
      },
      status: {
        title: "状态",
        summary: "从状态索引看当前进度：先看每个对象的状态事实，再看生命周期下一步。",
        nav: "状态对象"
      }
    },
    counts: {
      unit: "{count} 单元",
      scenario: "{count} 场景",
      rule: "{count} 规则",
      truth: "{count} Spec 文档",
      paths: "{count} 个路径或文件",
      objects: "{count} 个对象"
    },
    specflowSections: {
      unit: "单元",
      scenario: "场景",
      rule: "规则",
      truth: "Spec 文档",
      implementation: "实现路径",
      system: "全局规则",
      support: "支撑文件"
    },
    fallback: {
      statusUnknown: "状态未声明",
      nextStep: "下一步：{value}",
      none: "无",
      responsibilityUnknown: "职责未声明",
      undeclared: "未声明",
      rule: "规则",
      noObject: "暂无对象",
      cytoscapeMissing: "Cytoscape 未加载。"
    },
    statusBoard: {
      heading: "状态索引",
      description: "这些内容来自状态文件，但这里按对象卡片和表格展示，不直接显示 Markdown 原文。",
      sourceLabel: "来源：_status.md",
      metrics: {
        total: "登记对象",
        stable: "已有稳定设计",
        candidate: "正在确认",
        withImplementation: "已声明实现路径"
      },
      table: {
        object: "对象",
        layer: "当前层",
        next: "下一步",
        notes: "备注"
      },
      lifecycleHeading: "本轮生命周期进度",
      lifecycleDescription: "进度条表示当前这一轮。若本轮已完成，下一步会作为下一轮入口单独显示。",
      lifecycleAria: "{label} 生命周期位置",
      nextRoundEntry: "下一轮入口",
      nextRepairEntry: "修复入口"
    },
    todo: {
      empty: "暂无待处理动作。",
      emptyDetailTitle: "暂无待处理动作",
      emptyDetail: "状态索引里当前没有登记下一步动作。",
      boardHeading: "下一步动作",
      boardDescription: "每张卡片都来自 _status.md 的 Next Command。点击卡片查看需要打开的材料。",
      actionType: "动作类型",
      command: "命令",
      advanceEntry: "自动推进",
      copyAdvanceEntry: "复制自动推进入口",
      intent: "模式",
      materials: "可查看材料",
      references: "参考材料",
      implementation: "实现路径",
      notes: "原因",
      openMaterial: "打开材料",
      noMaterials: "暂无可读取材料",
      sourceLabels: {
        activeTruth: "当前 Spec",
        appendix: "附录",
        evidence: "证据",
        rule: "规则",
        checkResult: "检查结果",
        verifyResult: "验证结果",
        activePlan: "当前计划",
        status: "状态索引"
      },
      types: {
        stableVerify: "稳定复核",
        designCheck: "设计确认",
        plan: "开发计划",
        implementation: "实现执行",
        verify: "验证确认",
        promote: "沉淀基线",
        repairFork: "修复基线",
        fork: "开启变更轮次",
        new: "初始化 / 新建",
        other: "其他动作"
      },
      intents: {
        repair: "修复基线",
        change: "开启变更轮次"
      }
    },
    review: {
      empty: "暂无可查看 Spec。",
      emptyNav: "暂无 Spec 文档。",
      emptyDetailTitle: "暂无 Spec 文档",
      emptyDetail: "当前快照里没有可查看的 candidate 或 stable 主 Spec。",
      openSource: "打开 Spec 原文",
      fileType: "文档分组",
      object: "对应项目对象",
      reviewTarget: "查看说明",
      readingFocus: "查看重点",
      relationships: "相关关系",
      relationEmpty: "暂无相关关系快照。",
      progressTitle: "本轮进度",
      nextCommand: "下一步",
      noNextCommand: "当前没有登记下一步",
      copyNextCommand: "复制下一步命令",
      copied: "已复制",
      copyFailed: "复制失败",
      relation: {
        implementation: "实现路径",
        rule: "规则",
        ruleFile: "规则文件",
        bound: "绑定对象",
        appendix: "附录文件",
        evidence: "证据参考",
        stable: "稳定基线参考",
        mapping: "项目结构参考",
        system: "全局规则参考"
      },
      types: {
        candidate: "待确认 candidate",
        stable: "已确认 stable",
        stableRule: "已确认规则",
        capability: "单元设计",
        scenario: "端到端设计",
        rule: "规则",
        structure: "项目结构文件",
        system: "全局规则文件"
      },
      states: {
        candidate: "待确认",
        stable: "已确认",
        stableRule: "已确认"
      },
      targets: {
        candidate: "这是当前正在确认的 Spec，确认完成前不能当作正式基线。",
        stable: "这是已经确认的正式 Spec，可作为当前正式基线查看。",
        stableRule: "这是已经确认的共享规则 Spec，可作为当前正式规则查看。",
        capability: "整份文件是否正确表达该能力的当前设计或规则。",
        scenario: "整份文件是否正确表达从入口到最终结果的端到端链路。",
        rule: "整份文件是否正确表达这条规则及其复用边界。",
        structure: "整份文件是否正确表达当前项目结构、对象边界和路径归属。",
        system: "整份文件是否正确表达全仓库规则、默认选择和例外。"
      },
      focus: {
        candidate: "当前设计、边界、验收条件、附录、规则引用",
        stable: "正式设计、下一步动作、附录、证据、规则引用",
        stableRule: "规则正文、复用边界、绑定对象",
        capability: "责任边界、输入输出、错误处理、验收条件、规则引用",
        scenario: "入口、经过的能力、最终结果、失败处理、验证方式",
        rule: "复用对象、规则正文、绑定关系、是否仍是局部规则",
        structure: "能力列表、场景列表、规则列表、路径归属、支撑文件边界",
        system: "技术基线、默认选择、复用机制、禁止项、例外"
      }
    },
    lifecycle: {
      scenario_new: "创建新的端到端流程设计",
      scenario_stable_verify: "检查端到端流程是否仍符合已确认设计",
      scenario_check: "检查流程设计是否足够支撑验证",
      scenario_verify: "验证端到端流程",
      scenario_promote: "把流程确认结果沉淀为正式基线",
      scenario_fork: "从已确认流程开启新一轮设计",
      unit_init: "初始化能力真相",
      unit_stable_verify: "检查实现是否仍符合已确认设计",
      unit_new: "创建新的能力设计",
      unit_check: "检查设计是否足够支撑开发",
      unit_plan: "把设计整理成开发计划",
      unit_impl: "按计划实现",
      unit_verify: "验证实现是否符合设计",
      unit_promote: "把确认结果沉淀为正式基线",
      unit_fork: "从已确认基线开启新一轮设计"
    },
    lifecycleShort: {
      scenario_new: "新建",
      scenario_stable_verify: "稳定复核",
      scenario_check: "检查",
      scenario_verify: "验证",
      scenario_promote: "沉淀",
      scenario_fork: "开新轮",
      unit_init: "初始化",
      unit_stable_verify: "稳定复核",
      unit_new: "新建",
      unit_check: "检查",
      unit_plan: "计划",
      unit_impl: "实现",
      unit_verify: "验证",
      unit_promote: "沉淀",
      unit_fork: "开新轮"
    },
    inspector: {
      infoTab: "节点信息",
      truthTab: "Spec 文档",
      truthTitle: "Spec 文档",
      fields: {
        type: "类型",
        status: "状态",
        version: "版本",
        next: "下一步",
        responsibility: "职责",
        notes: "备注",
        file: "文件",
        connections: "连接",
        paths: "路径"
      },
      groups: {
        truth: "Spec 文档",
        implementation: "实现路径",
        rule: "规则",
        bound: "绑定对象",
        connected: "关联节点"
      }
    },
    docMode: {
      rendered: "渲染",
      raw: "原文"
    },
    source: {
      guideTitle: "导览",
      guideShow: "显示导览",
      guideHide: "隐藏导览",
      guideUnavailable: "无导览",
      noGuide: "暂无标题",
      emptyRendered: "选择一个 Spec 文档查看内容。",
      emptyRaw: "选择一个 Spec 文档查看原文。"
    },
    kind: {
      project_root: "仓库目录",
      project_path: "路径",
      project_area: "实现区域",
      repository_mapping: "项目结构文件",
      status_index: "状态索引",
      rule: "全局规则",
      truth_file: "Spec 文档"
    },
    frontmatter: {
      title: "元信息",
      undeclared: "未声明"
    }
  },
  en: {
    loading: "Loading",
    viewNavAria: "Reader views",
    graphLegendAria: "Graph node color legend",
    resizeAria: "Resize inspector panel",
    inspectorTabsAria: "Node inspector",
    docModeAria: "Spec document display mode",
    docGuideAria: "Spec document guide",
    refresh: "Refresh",
    language: {
      label: "Language",
      zh: "Chinese"
    },
    tabs: {
      todo: { title: "To Do", subtitle: "Next steps" },
      spec: { title: "Spec View", subtitle: "Documents" },
      status: { title: "Status", subtitle: "Progress" },
      project: { title: "Project", subtitle: "Repository paths" },
      specflow: { subtitle: "Governance layers" }
    },
    legend: {
      unit: {
        label: "Unit",
        tooltip: "A unit is an engineering responsibility that can be described, developed, and verified independently, such as agent, memory, or tool."
      },
      scenario: {
        label: "Scenario",
        tooltip: "A scenario is a complete trigger-to-result usage chain that shows how multiple responsibilities produce one outcome."
      },
      rule: {
        label: "Rule",
        tooltip: "A rule is reused by multiple units or scenarios so the same rule is not duplicated in different places."
      },
      truth: {
        label: "Spec document",
        tooltip: "A Spec document is accepted project text. Development and verification must follow these documents."
      },
      path: {
        label: "Implementation path",
        tooltip: "An implementation path is where code, configuration, or tooling files live."
      },
      system: {
        label: "Global rules",
        tooltip: "Global rules are repository-wide technical baselines, such as defaults, prohibitions, and global exceptions."
      }
    },
    views: {
      todo: {
        title: "To Do",
        summary: "Collects each object's next action from the status index. This is the unified entry: see what to do, then open the relevant material.",
        nav: "Next actions"
      },
      spec: {
        title: "Spec View",
        summary: "Shows candidate Specs that still need confirmation and stable Specs that are already accepted. Use To Do for the full next-action queue.",
        nav: "Spec documents"
      },
      project: {
        title: "Project Structure",
        summary: "Shows implementation locations from repository paths: which code or engineering paths are assigned to responsibility objects. SpecFlow's own Spec documents and support files are not shown here.",
        nav: "Directories",
        groups: {
          areas: "Implementation areas"
        }
      },
      specflow: {
        title: "SpecFlow",
        summary: "Shows governance layers: how global rules, repository mapping, status index, rules, units, and Spec documents are organized.",
        nav: "Objects"
      },
      status: {
        title: "Status",
        summary: "Shows current progress from the status index: object state facts first, then the next lifecycle step.",
        nav: "Status objects"
      }
    },
    counts: {
      unit: "{count} units",
      scenario: "{count} scenarios",
      rule: "{count} rules",
      truth: "{count} Spec documents",
      paths: "{count} paths or files",
      objects: "{count} objects"
    },
    specflowSections: {
      unit: "Units",
      scenario: "Scenarios",
      rule: "Rules",
      truth: "Spec documents",
      implementation: "Implementation paths",
      system: "Global rules",
      support: "Support files"
    },
    fallback: {
      statusUnknown: "Status not declared",
      nextStep: "Next: {value}",
      none: "None",
      responsibilityUnknown: "Responsibility not declared",
      undeclared: "Not declared",
      rule: "Rule",
      noObject: "No object",
      cytoscapeMissing: "Cytoscape is not loaded."
    },
    statusBoard: {
      heading: "Status Index",
      description: "This content comes from the status file, but is shown as object cards and tables instead of raw Markdown.",
      sourceLabel: "Source: _status.md",
      metrics: {
        total: "Registered objects",
        stable: "Stable designs",
        candidate: "In confirmation",
        withImplementation: "Implementation paths declared"
      },
      table: {
        object: "Object",
        layer: "Current layer",
        next: "Next",
        notes: "Notes"
      },
      lifecycleHeading: "Current Round Progress",
      lifecycleDescription: "The progress bar represents the current round. When the round is complete, the next command is shown separately as the next-round entry.",
      lifecycleAria: "{label} lifecycle position",
      nextRoundEntry: "Next-round entry",
      nextRepairEntry: "Repair entry"
    },
    todo: {
      empty: "No pending actions.",
      emptyDetailTitle: "No pending actions",
      emptyDetail: "The status index does not currently register a next action.",
      boardHeading: "Next Actions",
      boardDescription: "Each card comes from the Next Command field in _status.md. Select a card to see the material to open.",
      actionType: "Action type",
      command: "Command",
      advanceEntry: "Auto advance",
      copyAdvanceEntry: "Copy auto-advance entry",
      intent: "Mode",
      materials: "Readable material",
      references: "Reference material",
      implementation: "Implementation paths",
      notes: "Reason",
      openMaterial: "Open material",
      noMaterials: "No readable material",
      sourceLabels: {
        activeTruth: "Current Spec",
        appendix: "Appendix",
        evidence: "Evidence",
        rule: "Rule",
        checkResult: "Check result",
        verifyResult: "Verify result",
        activePlan: "Active plan",
        status: "Status index"
      },
      types: {
        stableVerify: "Stable verification",
        designCheck: "Design confirmation",
        plan: "Implementation plan",
        implementation: "Implementation",
        verify: "Verification",
        promote: "Promote baseline",
        repairFork: "Repair baseline",
        fork: "Start change round",
        new: "Initialize / create",
        other: "Other action"
      },
      intents: {
        repair: "Repair baseline",
        change: "Start change round"
      }
    },
    review: {
      empty: "No Specs to view.",
      emptyNav: "No Spec documents.",
      emptyDetailTitle: "No Spec documents",
      emptyDetail: "The current snapshot has no candidate or stable main Spec documents to view.",
      openSource: "Open Spec source",
      fileType: "Document group",
      object: "Project object",
      reviewTarget: "View note",
      readingFocus: "View focus",
      relationships: "Relationships",
      relationEmpty: "No relationship snapshot.",
      progressTitle: "Current round progress",
      nextCommand: "Next",
      noNextCommand: "No next command is registered",
      copyNextCommand: "Copy next command",
      copied: "Copied",
      copyFailed: "Copy failed",
      relation: {
        implementation: "Implementation paths",
        rule: "Rules",
        ruleFile: "Rule files",
        bound: "Bound objects",
        appendix: "Appendix files",
        evidence: "Evidence references",
        stable: "Stable baseline references",
        mapping: "Project structure reference",
        system: "Global rules reference"
      },
      types: {
        candidate: "Candidate to confirm",
        stable: "Accepted stable",
        stableRule: "Accepted rules",
        capability: "Unit design",
        scenario: "End-to-end design",
        rule: "Rule",
        structure: "Project structure file",
        system: "Global rules file"
      },
      states: {
        candidate: "To confirm",
        stable: "Accepted",
        stableRule: "Accepted"
      },
      targets: {
        candidate: "This Spec is still being confirmed and is not the formal baseline yet.",
        stable: "This Spec is already accepted and can be read as the current formal baseline.",
        stableRule: "This shared rule Spec is already accepted and can be read as the current formal rule.",
        capability: "Whether the whole file correctly expresses this capability's current design or rules.",
        scenario: "Whether the whole file correctly expresses the end-to-end chain from entry to final outcome.",
        rule: "Whether the whole file correctly expresses this rule and its reuse boundary.",
        structure: "Whether the whole file correctly expresses current project structure, object boundaries, and path ownership.",
        system: "Whether the whole file correctly expresses repository-wide constraints, defaults, and exceptions."
      },
      focus: {
        candidate: "Current design, boundaries, acceptance conditions, appendices, rule references",
        stable: "Formal design, next action, appendices, evidence, rule references",
        stableRule: "Rule body, reuse boundary, bound objects",
        capability: "Responsibility boundary, inputs and outputs, error handling, acceptance conditions, rule references",
        scenario: "Entry, participating capabilities, final outcome, failure handling, verification method",
        rule: "Reusing objects, rule body, binding relationships, whether it remains a local rule",
        structure: "Capability list, scenario list, rule list, path ownership, support-file boundary",
        system: "Technical baseline, defaults, reusable mechanisms, prohibitions, exceptions"
      }
    },
    lifecycle: {
      scenario_new: "Create a new end-to-end flow design",
      scenario_stable_verify: "Check whether the flow still matches the confirmed design",
      scenario_check: "Check whether the flow design is enough to support verification",
      scenario_verify: "Verify the end-to-end flow",
      scenario_promote: "Promote the confirmed flow result into the formal baseline",
      scenario_fork: "Start a new design round from a confirmed flow",
      unit_init: "Initialize capability truth",
      unit_stable_verify: "Check whether implementation still matches the confirmed design",
      unit_new: "Create a new capability design",
      unit_check: "Check whether the design is enough to support development",
      unit_plan: "Turn the design into an implementation plan",
      unit_impl: "Implement according to the plan",
      unit_verify: "Verify that implementation matches the design",
      unit_promote: "Promote the confirmed result into the formal baseline",
      unit_fork: "Start a new design round from a confirmed baseline"
    },
    lifecycleShort: {
      scenario_new: "New",
      scenario_stable_verify: "Stable check",
      scenario_check: "Check",
      scenario_verify: "Verify",
      scenario_promote: "Promote",
      scenario_fork: "Fork",
      unit_init: "Init",
      unit_stable_verify: "Stable check",
      unit_new: "New",
      unit_check: "Check",
      unit_plan: "Plan",
      unit_impl: "Implement",
      unit_verify: "Verify",
      unit_promote: "Promote",
      unit_fork: "Fork"
    },
    inspector: {
      infoTab: "Node Info",
      truthTab: "Spec Document",
      truthTitle: "Spec Document",
      fields: {
        type: "Type",
        status: "Status",
        version: "Version",
        next: "Next",
        responsibility: "Responsibility",
        notes: "Notes",
        file: "File",
        connections: "Connections",
        paths: "Paths"
      },
      groups: {
        truth: "Spec documents",
        implementation: "Implementation paths",
        rule: "Rules",
        bound: "Bound objects",
        connected: "Connected nodes"
      }
    },
    docMode: {
      rendered: "Rendered",
      raw: "Source"
    },
    source: {
      guideTitle: "Guide",
      guideShow: "Show Guide",
      guideHide: "Hide Guide",
      guideUnavailable: "No Guide",
      noGuide: "No headings",
      emptyRendered: "Select a Spec document to view its content.",
      emptyRaw: "Select a Spec document to view its source."
    },
    kind: {
      project_root: "Repository directory",
      project_path: "Path",
      project_area: "Implementation area",
      repository_mapping: "Repository mapping file",
      status_index: "Status index",
      rule: "Global rules",
      truth_file: "Spec document"
    },
    frontmatter: {
      title: "Metadata",
      undeclared: "Not declared"
    }
  }
};

const navPanel = document.getElementById("nav-panel");
const detailPanel = document.getElementById("detail-panel");
const graphView = document.getElementById("graph-view");
const viewSummary = document.getElementById("view-summary");
const projectMeta = document.getElementById("project-meta");
const sourcePath = document.getElementById("source-path");
const sourceContent = document.getElementById("source-content");
const sourceRendered = document.getElementById("source-rendered");
const docGuide = document.getElementById("doc-guide");
const docGuideToggle = document.getElementById("doc-guide-toggle");
const resizeBar = document.getElementById("resize-bar");
const infoTab = document.getElementById("info-tab");
const truthTab = document.getElementById("truth-tab");
const truthPanel = document.getElementById("truth-panel");
const languageSelect = document.getElementById("language-select");

document.getElementById("refresh-button").addEventListener("click", refreshReader);
languageSelect.value = currentLanguage;
languageSelect.addEventListener("change", () => setLanguage(languageSelect.value));
document.querySelectorAll(".tab").forEach((button) => {
  button.addEventListener("click", () => {
    currentView = button.dataset.view;
    document.querySelectorAll(".tab").forEach((item) => item.classList.toggle("active", item === button));
    render();
  });
});

document.querySelectorAll("[data-inspector-tab]").forEach((button) => {
  button.addEventListener("click", () => {
    if (button.classList.contains("hidden")) return;
    setInspectorTab(button.dataset.inspectorTab);
  });
});

document.querySelectorAll("[data-doc-mode]").forEach((button) => {
  button.addEventListener("click", () => setDocMode(button.dataset.docMode));
});
docGuideToggle.addEventListener("click", () => setDocGuideOpen(!docGuideOpen));

resizeBar.addEventListener("pointerdown", startInspectorResize);

function readStoredLanguage() {
  try {
    const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (SUPPORTED_LANGUAGES.includes(stored)) return stored;
  } catch {
    return "zh-CN";
  }
  return "zh-CN";
}

function setLanguage(language) {
  currentLanguage = SUPPORTED_LANGUAGES.includes(language) ? language : "zh-CN";
  languageSelect.value = currentLanguage;
  document.documentElement.lang = currentLanguage;
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, currentLanguage);
  } catch {
    // Browser storage can be unavailable in restricted contexts.
  }
  applyStaticText();
  render();
}

function t(key, params = {}) {
  const primary = lookupTranslation(TRANSLATIONS[currentLanguage], key);
  const fallback = lookupTranslation(TRANSLATIONS["zh-CN"], key);
  const template = primary ?? fallback ?? key;
  return String(template).replace(/\{([A-Za-z0-9_]+)\}/g, (_, name) => {
    return Object.prototype.hasOwnProperty.call(params, name) ? params[name] : "";
  });
}

function lookupTranslation(source, key) {
  return String(key).split(".").reduce((value, part) => {
    if (value && Object.prototype.hasOwnProperty.call(value, part)) return value[part];
    return undefined;
  }, source);
}

function applyStaticText() {
  document.documentElement.lang = currentLanguage;
  document.querySelectorAll("[data-i18n]").forEach((element) => {
    element.textContent = t(element.dataset.i18n);
  });
  document.querySelectorAll("[data-i18n-attr]").forEach((element) => {
    element.dataset.i18nAttr.split(";").forEach((item) => {
      const [attribute, key] = item.split(":");
      if (attribute && key) element.setAttribute(attribute, t(key));
    });
  });
  if (!sourcePath.textContent) {
    sourceRendered.textContent = t("source.emptyRendered");
    sourceContent.textContent = t("source.emptyRaw");
  }
  renderDocGuide(activeSourceHeadings);
  bindDocGuideLinks();
  updateDocGuideToggle();
}

async function loadSnapshot() {
  if (snapshotRequestInFlight) return;
  snapshotRequestInFlight = true;
  try {
    const response = await fetch("/api/snapshot");
    const nextSnapshot = await response.json();
    const nextSignature = snapshotSignature(nextSnapshot);
    if (!snapshot || nextSignature !== snapshotDataSignature) {
      snapshot = nextSnapshot;
      snapshotDataSignature = nextSignature;
      render();
    }
  } finally {
    snapshotRequestInFlight = false;
  }
}

async function refreshReader() {
  const openPath = sourcePath.textContent.trim();
  await loadSnapshot();
  if (openPath && isReadableOriginalPath(openPath)) {
    await openSource(openPath, { activate: activeInspectorTab === "truth" });
  }
}

function snapshotSignature(value) {
  if (!value) return "";
  const comparable = { ...value };
  delete comparable.version;
  delete comparable.generated_at;
  return JSON.stringify(comparable);
}

function render() {
  if (!snapshot) return;
  document.body.classList.toggle("todo-view-active", currentView === "todo");
  document.body.classList.toggle("spec-view-active", currentView === "spec");
  document.body.classList.toggle("status-view-active", currentView === "status");
  const objects = list(snapshot.objects);
  projectMeta.textContent = `${snapshot.project.repo_root} · version ${snapshot.version} · ${t("counts.objects", { count: objects.length })}`;
  const graph = graphForCurrentView();
  if (!selectedNodeID || !nodeExistsForGraph(selectedNodeID, graph)) {
    selectedNodeID = firstNodeIDForView(graph.nodes);
  }
  renderViewSummary();
  renderNav();
  renderGraph();
  renderDetailForNode(selectedNodeID);
}

function renderViewSummary() {
  const viewKey = `views.${currentView}`;
  viewSummary.innerHTML = `
    <div>
      <h2>${escapeHTML(t(`${viewKey}.title`))}</h2>
      <p>${escapeHTML(t(`${viewKey}.summary`))}</p>
    </div>
    <div class="view-counts">
      <span>${escapeHTML(t("counts.unit", { count: snapshot.project.unit_count || 0 }))}</span>
      <span>${escapeHTML(t("counts.scenario", { count: snapshot.project.scenario_count || 0 }))}</span>
      <span>${escapeHTML(t("counts.rule", { count: snapshot.project.rule_count || 0 }))}</span>
      <span>${escapeHTML(t("counts.truth", { count: snapshot.project.truth_file_count || 0 }))}</span>
    </div>
  `;
}

function renderNav() {
  navPanel.innerHTML = "";
  const navTitle = document.createElement("div");
  navTitle.className = "nav-title";
  navTitle.textContent = t(`views.${currentView}.nav`);
  navPanel.appendChild(navTitle);

  const diagnostics = list(snapshot.diagnostics);
  if (diagnostics.length > 0) {
    diagnostics.forEach((diagnostic) => {
      const div = document.createElement("div");
      div.className = "diagnostic";
      div.textContent = `${diagnostic.severity}: ${diagnostic.message}`;
      navPanel.appendChild(div);
    });
  }

  if (currentView === "project") {
    const graph = graphForCurrentView();
    const objects = graph.nodes
      .filter((node) => (node.group === "unit" || node.group === "scenario" || node.group === "rule") && list(node.raw_paths).length > 0)
      .sort(byLabel);
    objects.forEach((node) => {
      const button = document.createElement("button");
      button.className = node.id === selectedNodeID ? "nav-item active" : "nav-item";
      button.type = "button";
      button.innerHTML = `<strong>${escapeHTML(node.label)}</strong><span>${escapeHTML(t("counts.paths", { count: list(node.raw_paths).length }))}</span>`;
      button.addEventListener("click", () => focusNode(node.id));
      navPanel.appendChild(button);
    });
    return;
  }

  if (currentView === "specflow") {
    renderSpecflowNav();
    return;
  }

  if (currentView === "todo") {
    renderTodoNav();
    return;
  }

  if (currentView === "spec") {
    renderReviewNav();
    return;
  }

  if (currentView === "status") {
    objectsForView().forEach((object) => {
      const button = document.createElement("button");
      button.className = objectNodeID(object) === selectedNodeID ? "nav-item active" : "nav-item";
      button.type = "button";
      button.innerHTML = `<strong>${escapeHTML(object.label)}</strong><span>${escapeHTML(navSubtitle(object))}</span>`;
      button.addEventListener("click", () => focusObject(object));
      navPanel.appendChild(button);
    });
    return;
  }

  const objects = objectsForView();
  objects.forEach((object) => {
    const button = document.createElement("button");
    button.className = objectNodeID(object) === selectedNodeID ? "nav-item active" : "nav-item";
    button.type = "button";
    button.innerHTML = `<strong>${escapeHTML(object.label)}</strong><span>${escapeHTML(navSubtitle(object))}</span>`;
    button.addEventListener("click", () => focusObject(object));
    navPanel.appendChild(button);
  });
}

function renderSpecflowNav() {
  const graph = graphForSpecflowView();
  const objects = list(snapshot.objects);
  const units = objects.filter((item) => item.kind === "unit").sort(byLabel);
  const scenarios = objects.filter((item) => item.kind === "scenario").sort(byLabel);
  const rules = objects.filter((item) => item.kind === "rule").sort(byLabel);
  const truthNodes = graph.nodes.filter((node) => node.group === "truth").sort(byLabel);
  const implementationNodes = graph.nodes.filter((node) => node.group === "implementation").sort(byLabel);
  const systemNodes = graph.nodes.filter((node) => node.group === "__unused_rule_group__").sort(byLabel);
  const supportNodes = graph.nodes.filter((node) => node.group === "support").sort(byLabel);

  const sections = [
    { key: "unit", type: "objects", items: units },
    { key: "scenario", type: "objects", items: scenarios },
    { key: "rule", type: "objects", items: rules },
    { key: "truth", type: "nodes", items: truthNodes },
    { key: "implementation", type: "nodes", items: implementationNodes },
    { key: "system", type: "nodes", items: systemNodes },
    { key: "support", type: "nodes", items: supportNodes }
  ].filter((section) => section.items.length > 0);
  if (!sections.some((section) => section.key === activeSpecflowNavGroup)) {
    activeSpecflowNavGroup = (sections[0] || {}).key || "unit";
  }
  sections.forEach((section) => {
    if (section.type === "objects") {
      renderObjectNavSection(section.key, section.items);
      return;
    }
    renderNodeNavSection(section.key, section.items);
  });
}

function renderReviewNav() {
  const items = reviewItems();
  const sections = reviewTypeOrder()
    .map((type) => ({ type, items: items.filter((item) => item.reviewType === type) }))
    .filter((section) => section.items.length > 0);
  if (sections.length === 0) {
    const empty = document.createElement("div");
    empty.className = "nav-empty";
    empty.textContent = t("review.emptyNav");
    navPanel.appendChild(empty);
    return;
  }
  if (!sections.some((section) => section.type === activeReviewNavGroup)) {
    activeReviewNavGroup = (sections[0] || {}).type || "capability";
  }
  sections.forEach((section) => renderReviewNavSection(section.type, section.items));
}

function renderReviewNavSection(type, items) {
  const expanded = type === activeReviewNavGroup;
  const section = document.createElement("section");
  section.className = expanded ? "nav-section expanded" : "nav-section";

  const header = document.createElement("button");
  header.className = "nav-section-title";
  header.type = "button";
  header.setAttribute("aria-expanded", String(expanded));
  header.innerHTML = `<span>${escapeHTML(reviewTypeLabel(type))}</span><em>${items.length}</em>`;
  header.addEventListener("click", () => {
    activeReviewNavGroup = type;
    renderNav();
  });
  section.appendChild(header);

  if (expanded) {
    items.forEach((item) => {
      const button = document.createElement("button");
      button.className = item.id === selectedNodeID ? "nav-item active" : "nav-item";
      button.type = "button";
      button.innerHTML = `<strong>${escapeHTML(item.fileLabel)}</strong><span>${escapeHTML(reviewNavSubtitle(item))}</span>`;
      button.addEventListener("click", () => focusReviewItem(item.id));
      section.appendChild(button);
    });
  }
  navPanel.appendChild(section);
}

function renderObjectNavSection(sectionKey, objects) {
  if (!objects || objects.length === 0) return;
  const section = createNavSection(sectionKey, objects.length);
  if (sectionKey !== activeSpecflowNavGroup) {
    navPanel.appendChild(section);
    return;
  }
  objects.forEach((object) => {
    const button = document.createElement("button");
    button.className = objectNodeID(object) === selectedNodeID ? "nav-item active" : "nav-item";
    button.type = "button";
    button.innerHTML = `<strong>${escapeHTML(object.label)}</strong><span>${escapeHTML(navSubtitle(object))}</span>`;
    button.addEventListener("click", () => focusObject(object));
    section.appendChild(button);
  });
  navPanel.appendChild(section);
}

function renderNodeNavSection(sectionKey, nodes) {
  if (!nodes || nodes.length === 0) return;
  const section = createNavSection(sectionKey, nodes.length);
  if (sectionKey !== activeSpecflowNavGroup) {
    navPanel.appendChild(section);
    return;
  }
  nodes.forEach((node) => {
    const button = document.createElement("button");
    button.className = node.id === selectedNodeID ? "nav-item active" : "nav-item";
    button.type = "button";
    button.innerHTML = `<strong>${escapeHTML(compactLabel(node))}</strong><span>${escapeHTML(node.source && node.source.path ? node.source.path : labelForKind(node.kind))}</span>`;
    button.addEventListener("click", () => focusNode(node.id));
    section.appendChild(button);
  });
  navPanel.appendChild(section);
}

function createNavSection(sectionKey, count) {
  const section = document.createElement("section");
  const expanded = sectionKey === activeSpecflowNavGroup;
  section.className = expanded ? "nav-section expanded" : "nav-section";
  const header = document.createElement("button");
  header.className = "nav-section-title";
  header.type = "button";
  header.setAttribute("aria-expanded", String(expanded));
  header.innerHTML = `<span>${escapeHTML(t(`specflowSections.${sectionKey}`))}</span><em>${count}</em>`;
  header.addEventListener("click", () => {
    activeSpecflowNavGroup = sectionKey;
    renderNav();
  });
  section.appendChild(header);
  return section;
}

function objectsForView() {
  const objects = list(snapshot.objects);
  if (currentView === "project") {
    return objects.filter((item) => item.kind === "unit" || item.kind === "scenario");
  }
  if (currentView === "specflow") {
    return objects.filter((item) => item.kind === "rule").concat(objects.filter((item) => item.kind === "unit"));
  }
  if (currentView === "status") {
    return objects.filter((item) => item.kind === "unit" || item.kind === "scenario");
  }
  return objects;
}

function navSubtitle(object) {
  if (currentView === "status") {
    return `${object.human_state || object.layer || t("fallback.statusUnknown")} · ${t("fallback.nextStep", { value: object.next_label || object.next_command || t("fallback.none") })}`;
  }
  if (currentView === "specflow") {
    return object.responsibility || t("fallback.responsibilityUnknown");
  }
  if (object.kind === "rule") return object.responsibility || t("fallback.rule");
  return `${object.human_state || object.kind} · ${t("fallback.nextStep", { value: object.next_label || t("fallback.none") })}`;
}

function focusObject(object) {
  selectedNodeID = objectNodeID(object);
  renderNav();
  renderDetailForNode(selectedNodeID);
  focusGraphNode(selectedNodeID, 1.05);
}

function focusNode(nodeID) {
  selectedNodeID = nodeID;
  renderNav();
  renderDetailForNode(nodeID);
  focusGraphNode(nodeID, 1.05);
}

function focusGraphNode(nodeID, zoom) {
  if (!cy || !nodeID) return;
  const node = cy.getElementById(nodeID);
  if (node.length > 0) {
    cy.elements().removeClass("selected");
    node.addClass("selected");
    cy.animate({ center: { eles: node }, zoom }, { duration: 250 });
  }
}

function renderGraph() {
  if (currentView === "todo") {
    if (cy) {
      cy.destroy();
      cy = null;
    }
    renderTodoBoard();
    return;
  }
  if (currentView === "spec") {
    if (cy) {
      cy.destroy();
      cy = null;
    }
    renderReviewBoard();
    return;
  }
  if (currentView === "status") {
    if (cy) {
      cy.destroy();
      cy = null;
    }
    renderStatusBoard();
    return;
  }
  if (typeof cytoscape !== "function") {
    graphView.textContent = t("fallback.cytoscapeMissing");
    return;
  }
  if (cy) {
    cy.destroy();
    cy = null;
  }
  graphView.innerHTML = "";
  const graph = graphForCurrentView();
  const positions = readablePositions(graph.nodes, graph.edges);
  const elements = [];
  graph.nodes.forEach((node) => {
    elements.push({
      data: {
        id: node.id,
        label: compactLabel(node),
        kind: node.kind,
        group: node.group,
        source: node.source || null
      },
      position: positions[node.id] || { x: 100, y: 100 }
    });
  });
  graph.edges.forEach((edge) => {
    elements.push({
      data: {
        id: edge.id,
        source: edge.from,
        target: edge.to,
        label: edgeLabel(edge.kind),
        kind: edge.kind,
        sourceRef: edge.source || null
      }
    });
  });
  cy = cytoscape({
    container: graphView,
    elements,
    style: [
      { selector: "node", style: {
        "label": "data(label)",
        "text-wrap": "wrap",
        "text-max-width": 150,
        "font-size": 12,
        "color": "#1f2933",
        "text-halign": "right",
        "text-valign": "center",
        "text-margin-x": 11,
        "text-margin-y": 0,
        "background-color": colorForGroup,
        "border-width": 1,
        "border-color": "#ffffff",
        "width": nodeSize,
        "height": nodeSize,
        "text-background-color": "#ffffff",
        "text-background-opacity": 0.86,
        "text-background-padding": 2
      }},
      { selector: 'node[group = "implementation"]', style: {
        "text-halign": "left",
        "text-margin-x": -11
      }},
      { selector: "edge", style: {
        "curve-style": "bezier",
        "target-arrow-shape": "triangle",
        "target-arrow-color": "#94a3b8",
        "line-color": "#94a3b8",
        "width": edgeWidth,
        "label": "",
        "opacity": 0.7
      }},
      { selector: ".selected", style: {
        "border-width": 5,
        "border-color": "#2563eb"
      }},
      { selector: ".connected", style: {
        "line-color": "#2563eb",
        "target-arrow-color": "#2563eb",
        "opacity": 1,
        "width": 2.4
      }}
    ],
    minZoom: 0.35,
    maxZoom: 2.2,
    wheelSensitivity: 1.1,
    layout: { name: "preset", fit: false }
  });
  cy.ready(() => {
    focusGraphNode(selectedNodeID || firstNodeIDForView(graph.nodes), 0.85);
  });
  cy.on("tap", "node", (event) => {
    const data = event.target.data();
    selectedNodeID = data.id;
    highlightConnected(data.id);
    renderNav();
    renderDetailForNode(data.id);
  });
  cy.on("tap", "edge", (event) => {
    const source = event.target.data("sourceRef");
    if (source && isReadableOriginalPath(source.path)) openSource(source.path);
  });
}

function graphForCurrentView() {
  if (currentView === "todo") return graphForTodoView();
  if (currentView === "spec") return graphForReviewView();
  if (currentView === "project") return graphForProjectView();
  if (currentView === "specflow") return graphForSpecflowView();
  if (currentView === "status") return graphForStatusView();

  const nodes = list(snapshot.nodes);
  const edges = list(snapshot.edges);
  return { nodes, edges };
}

function graphForTodoView() {
  return {
    nodes: todoItems().map((item) => ({
      id: item.id,
      kind: item.object.kind,
      label: item.objectLabel,
      group: item.type,
      source: firstSourceRef(item.sources)
    })),
    edges: []
  };
}

function graphForStatusView() {
  return {
    nodes: objectsForView().map((object) => ({
      id: objectNodeID(object),
      kind: object.kind,
      label: object.label,
      group: object.kind,
      source: firstSourceRef(object.sources)
    })),
    edges: []
  };
}

function graphForReviewView() {
  return {
    nodes: reviewItems().map((item) => ({
      id: item.id,
      kind: item.reviewType,
      label: item.fileLabel,
      group: item.reviewType,
      source: item.source
    })),
    edges: []
  };
}

function graphForProjectView() {
  const nodesByID = new Map();
  const edges = [];
  const rootsByID = new Map();
  const areasByID = new Map();
  const addNode = (node) => {
    if (!nodesByID.has(node.id)) nodesByID.set(node.id, node);
  };
  const updateNode = (node) => {
    nodesByID.set(node.id, { ...(nodesByID.get(node.id) || {}), ...node });
  };
  const addEdge = (edge) => {
      if (!edges.some((item) => item.id === edge.id)) edges.push(edge);
  };

  list(snapshot.objects).forEach((object) => {
    const implementationPaths = list(object.implementation_paths);
    if (implementationPaths.length === 0) return;
    const objectID = objectNodeID(object);
    addNode({
      id: objectID,
      kind: object.kind,
      label: object.label,
      group: object.kind === "rule" ? "rule" : object.kind,
      source: firstSourceRef(object.sources),
      raw_paths: implementationPaths
    });
    implementationPaths.forEach((ref) => {
      const pathID = addProjectArea(addNode, addEdge, rootsByID, areasByID, ref);
      addEdge({
        id: `${pathID}->${objectID}`,
        from: pathID,
        to: objectID,
        kind: "maps_to",
        label: "maps to",
        source: ref
      });
    });
  });

  areasByID.forEach((area) => {
    updateNode({
      id: area.id,
      label: `${area.aggregate_path_label} · ${t("counts.paths", { count: area.raw_paths.length })}`,
      aggregate_path_count: area.raw_paths.length,
      raw_paths: area.raw_paths
    });
  });
  rootsByID.forEach((root) => {
    updateNode({
      id: root.id,
      aggregate_path_count: root.raw_paths.length,
      raw_paths: root.raw_paths
    });
  });

  return { nodes: [...nodesByID.values()], edges };
}

function addProjectArea(addNode, addEdge, rootsByID, areasByID, ref) {
  if (!ref || !ref.path) return "";
  const root = rootForImplementationPath(ref.path);
  if (!rootsByID.has(root.id)) {
    root.raw_paths = [];
    rootsByID.set(root.id, root);
    addNode(root);
  }
  rootsByID.get(root.id).raw_paths.push(ref);

  const areaPath = projectAreaForImplementationPath(ref.path);
  const pathID = `project_area:${areaPath}`;
  if (!areasByID.has(pathID)) {
    const area = {
      id: pathID,
      kind: "project_area",
      label: areaPath,
      group: "implementation",
      source: sourceForImplementationRef(ref),
      aggregate_path_label: areaPath,
      aggregate_path_count: 0,
      raw_paths: []
    };
    areasByID.set(pathID, area);
    addNode(area);
  }
  areasByID.get(pathID).raw_paths.push(ref);
  addEdge({
    id: `${root.id}->${pathID}`,
    from: root.id,
    to: pathID,
    kind: "contains",
    label: "contains",
    source: ref
  });
  return pathID;
}

function projectAreaForImplementationPath(rawPath) {
  const path = String(rawPath || "").replaceAll("\\", "/").replace(/\/\*\*$/, "/");
  if (path.endsWith("/")) return path;
  const segments = path.split("/");
  if (segments.length <= 1) return path;
  segments.pop();
  return `${segments.join("/")}/`;
}

function rootForImplementationPath(path) {
  const firstSegment = String(path).split("/")[0] || "unknown";
  const label = path.includes("/") ? `${firstSegment}/` : firstSegment;
  return { id: `root:${label}`, kind: "project_root", label, group: "root" };
}

function graphForSpecflowView() {
  const nodesByID = new Map();
  const edges = [];
  const addNode = (node) => {
    if (!nodesByID.has(node.id)) nodesByID.set(node.id, node);
  };
  const addEdge = (edge) => {
    if (!edges.some((item) => item.id === edge.id)) edges.push(edge);
  };

  addNode({ id: "rule:baseline", kind: "rule", label: t("kind.rule"), group: "rule", source: { path: snapshot.project.rule_baseline_file } });
  addNode({ id: "support:repository_mapping", kind: "repository_mapping", label: t("kind.repository_mapping"), group: "support", source: { path: snapshot.project.mapping_file } });
  addNode({ id: "support:status", kind: "status_index", label: t("kind.status_index"), group: "support", source: { path: snapshot.project.status_file } });
  addEdge({ id: "rule:baseline->support:repository_mapping", from: "rule:baseline", to: "support:repository_mapping", kind: "constrains", label: "constrains", source: { path: snapshot.project.rule_baseline_file } });

  list(snapshot.objects).forEach((object) => {
    const objectID = objectNodeID(object);
    addNode({
      id: objectID,
      kind: object.kind,
      label: object.label,
      group: object.kind === "rule" ? "rule" : object.kind,
      source: firstSourceRef(object.sources)
    });
    addEdge({ id: `support:repository_mapping->${objectID}`, from: "support:repository_mapping", to: objectID, kind: "declares", label: "declares", source: { path: snapshot.project.mapping_file } });
    if (object.kind === "unit" || object.kind === "scenario") {
      addEdge({ id: `support:status->${objectID}`, from: "support:status", to: objectID, kind: "tracks_state", label: "tracks state", source: { path: snapshot.project.status_file } });
    }
    list(object.truth_paths).forEach((truth) => {
      const fileNode = `file:${truth.path}`;
      addNode({ id: fileNode, kind: "truth_file", label: truth.path.split("/").pop(), group: "truth", source: truth });
      addEdge({ id: `${objectID}->${fileNode}`, from: objectID, to: fileNode, kind: "described_by", label: "described by", source: truth });
    });
    list(object.rule_refs).forEach((ruleID) => {
      const ruleNode = `rule:${ruleID}`;
      addNode({ id: ruleNode, kind: "rule", label: ruleID, group: "rule" });
      addEdge({ id: `${objectID}->${ruleNode}`, from: objectID, to: ruleNode, kind: "uses_rule", label: "uses rule", source: firstSourceRef(object.sources) });
    });
    list(object.bound_objects).forEach((bound) => {
      if (object.kind !== "rule") return;
      addEdge({ id: `${objectID}->${bound}`, from: objectID, to: bound, kind: "bound_to", label: "bound to", source: firstSourceRef(object.sources) });
    });
  });

  return { nodes: [...nodesByID.values()], edges };
}

function firstSourceRef(sources) {
  return list(sources).find((source) => source && source.path) || null;
}

function readablePositions(nodes, edges) {
  if (currentView === "project") return projectPositions(nodes, edges);
  if (currentView === "specflow") return specflowPositions(nodes, edges);
  return relationshipPositions(nodes, edges);
}

function projectPositions(nodes, edges) {
  const positions = {};
  const roots = nodes.filter((node) => node.group === "root").sort(byLabel);
  const areas = nodes.filter((node) => node.kind === "project_area").sort(byLabel);
  const objects = nodes
    .filter((node) => node.kind !== "project_area" && node.group !== "root")
    .sort(byLabel);
  const top = 80;
  const areaGap = 68;
  const rootGap = 150;
  const objectGap = 116;
  const rootX = 120;
  const areaX = 520;
  const objectX = 860;

  objects.forEach((node, index) => {
    positions[node.id] = { x: objectX, y: top + index * objectGap };
  });

  areas.forEach((node, index) => {
    const owners = edges
      .filter((edge) => edge.from === node.id && positions[edge.to])
      .map((edge) => positions[edge.to].y);
    positions[node.id] = { x: areaX, y: owners.length > 0 ? average(owners) : top + index * areaGap };
  });
  distributeColumn(areas, positions, areaX, top, areaGap);

  roots.forEach((node, index) => {
    const children = edges
      .filter((edge) => edge.from === node.id && positions[edge.to])
      .map((edge) => positions[edge.to].y);
    positions[node.id] = { x: rootX, y: children.length > 0 ? average(children) : top + index * rootGap };
  });
  distributeColumn(roots, positions, rootX, top, rootGap);
  return positions;
}

function specflowPositions(nodes, edges) {
  const positions = {};
  const groups = {
    system: nodes.filter((node) => node.group === "support"),
    rule: nodes.filter((node) => node.group === "rule"),
    domain: nodes.filter((node) => node.group === "unit" || node.group === "scenario"),
    truth: nodes.filter((node) => node.group === "truth")
  };
  const x = { system: 120, rule: 390, domain: 650, truth: 940 };
  const top = 90;

  groups.system.sort(byLabel).forEach((node, index) => {
    positions[node.id] = { x: x.system, y: top + index * 110 };
  });
  groups.rule.sort(byLabel).forEach((node, index) => {
    positions[node.id] = { x: x.rule, y: top + index * 118 };
  });
  groups.domain.sort(byLabel).forEach((node, index) => {
    positions[node.id] = { x: x.domain, y: top + index * 118 };
  });
  positionChildGroup(groups.truth, edges, positions, x.truth, top, 72);
  distributeColumn(groups.truth, positions, x.truth, top, 60);
  return positions;
}

function relationshipPositions(nodes, edges) {
  const positions = {};
  const groups = {
    rule: nodes.filter((node) => node.group === "rule"),
    domain: nodes.filter((node) => node.group === "unit" || node.group === "scenario"),
    truth: nodes.filter((node) => node.group === "truth"),
    implementation: nodes.filter((node) => node.group === "implementation"),
    system: nodes.filter((node) => node.group === "support")
  };
  const x = { rule: 120, system: 120, domain: 430, truth: 760, implementation: 1100 };
  const rowGap = 135;
  const top = 90;

  groups.domain.sort(byLabel).forEach((node, index) => {
    positions[node.id] = { x: x.domain, y: top + index * rowGap };
  });
  groups.system.sort(byLabel).forEach((node, index) => {
    positions[node.id] = { x: x.system, y: top + index * rowGap };
  });

  groups.rule.sort(byLabel).forEach((node, index) => {
    const boundTargets = edges
      .filter((edge) => edge.from === node.id && positions[edge.to])
      .map((edge) => positions[edge.to].y);
    const y = boundTargets.length > 0 ? average(boundTargets) : top + index * rowGap;
    positions[node.id] = { x: x.rule, y };
  });

  distributeColumn(groups.rule.concat(groups.system), positions, x.rule, top, 112);
  positionChildGroup(groups.truth, edges, positions, x.truth, top, 72);
  positionChildGroup(groups.implementation, edges, positions, x.implementation, top, 82);

  nodes.forEach((node, index) => {
    if (!positions[node.id]) {
      positions[node.id] = { x: 430, y: top + index * rowGap };
    }
  });
  return positions;
}

function positionChildGroup(nodes, edges, positions, x, fallbackTop, gap) {
  const childrenByParent = new Map();
  nodes.sort(byLabel).forEach((node) => {
    const parentEdge = edges.find((edge) => edge.to === node.id && positions[edge.from]);
    const parent = parentEdge ? parentEdge.from : "";
    if (!childrenByParent.has(parent)) childrenByParent.set(parent, []);
    childrenByParent.get(parent).push(node);
  });
  let nextY = fallbackTop;
  [...childrenByParent.entries()]
    .map(([parent, children]) => {
      const parentY = parent ? positions[parent].y : fallbackTop;
      return { parent, children, parentY };
    })
    .sort((left, right) => left.parentY - right.parentY || String(left.parent).localeCompare(String(right.parent)))
    .forEach(({ children, parentY }) => {
      const centeredStart = parentY - ((children.length - 1) * gap) / 2;
      const startY = Math.max(centeredStart, nextY);
      children.forEach((node, index) => {
        positions[node.id] = { x, y: startY + index * gap };
      });
      nextY = startY + children.length * gap;
    });
}

function distributeColumn(nodes, positions, x, fallbackTop, gap) {
  let nextY = fallbackTop;
  nodes
    .filter((node) => positions[node.id])
    .sort((left, right) => positions[left.id].y - positions[right.id].y || byLabel(left, right))
    .forEach((node) => {
      const y = Math.max(positions[node.id].y, nextY);
      positions[node.id] = { x, y };
      nextY = y + gap;
    });
}

function highlightConnected(nodeID) {
  if (!cy) return;
  cy.elements().removeClass("selected connected");
  const node = cy.getElementById(nodeID);
  node.addClass("selected");
  node.connectedEdges().addClass("connected");
}

function colorForGroup(ele) {
  const group = ele.data("group");
  if (group === "unit") return "#2563eb";
  if (group === "scenario") return "#db2777";
  if (group === "rule") return "#0f766e";
  if (group === "truth") return "#7c3aed";
  if (group === "implementation") return "#b45309";
  if (group === "root") return "#0f172a";
  if (group === "support") return "#64748b";
  return "#475569";
}

function objectFromNode(id) {
  if (!id.includes(":")) return null;
  const [kind, objectID] = id.split(":", 2);
  const objectKind = kind === "rule" ? "rule" : kind;
  return list(snapshot.objects).find((item) => item.kind === objectKind && item.id === objectID);
}

function objectNodeID(object) {
  if (!object) return null;
  return `${object.kind === "rule" ? "rule" : object.kind}:${object.id}`;
}

function nodeExistsForGraph(nodeID, graph) {
  return list(graph.nodes).some((node) => node.id === nodeID);
}

function firstNodeIDForView(nodes) {
  if (currentView === "todo") {
    return (nodes[0] || {}).id || null;
  }
  if (currentView === "spec") {
    return (nodes[0] || {}).id || null;
  }
  if (currentView === "project") {
    const objectNode = nodes.find((node) => (node.group === "unit" || node.group === "scenario" || node.group === "rule") && list(node.raw_paths).length > 0);
    return (objectNode || nodes[0] || {}).id || null;
  }
  if (currentView === "specflow") {
    const supportNode = nodes.find((node) => node.id === "rule:baseline");
    return (supportNode || nodes[0] || {}).id || null;
  }
  if (currentView === "status") {
    return (nodes[0] || {}).id || null;
  }
  const domainNode = nodes.find((node) => node.group === "unit" || node.group === "scenario");
  return (domainNode || nodes[0] || {}).id || null;
}

function renderStatusBoard() {
  const objects = objectsForView();
  const overview = statusOverview(objects);
  graphView.innerHTML = `
    <div class="status-board">
      <section class="status-section">
        <div class="status-section-heading">
          <div>
            <h3>${escapeHTML(t("statusBoard.heading"))}</h3>
            <p>${escapeHTML(t("statusBoard.description"))}</p>
          </div>
          ${renderSourceButton(snapshot.project.status_file, t("statusBoard.sourceLabel"))}
        </div>
        <div class="metric-grid">
          <div class="metric"><strong>${overview.total}</strong><span>${escapeHTML(t("statusBoard.metrics.total"))}</span></div>
          <div class="metric"><strong>${overview.stable}</strong><span>${escapeHTML(t("statusBoard.metrics.stable"))}</span></div>
          <div class="metric"><strong>${overview.candidate}</strong><span>${escapeHTML(t("statusBoard.metrics.candidate"))}</span></div>
          <div class="metric"><strong>${overview.withImplementation}</strong><span>${escapeHTML(t("statusBoard.metrics.withImplementation"))}</span></div>
        </div>
        <div class="status-table-wrap">
          <table class="status-table">
            <thead>
              <tr>
                <th>${escapeHTML(t("statusBoard.table.object"))}</th>
                <th>${escapeHTML(t("statusBoard.table.layer"))}</th>
                <th>Stable</th>
                <th>Candidate</th>
                <th>${escapeHTML(t("statusBoard.table.next"))}</th>
                <th>${escapeHTML(t("statusBoard.table.notes"))}</th>
              </tr>
            </thead>
            <tbody>${objects.map(renderStatusRow).join("")}</tbody>
          </table>
        </div>
      </section>

      <section class="status-section">
        <div class="status-section-heading">
          <div>
            <h3>${escapeHTML(t("statusBoard.lifecycleHeading"))}</h3>
            <p>${escapeHTML(t("statusBoard.lifecycleDescription"))}</p>
          </div>
        </div>
        <div class="lifecycle-list">${objects.map(renderLifecycleCard).join("")}</div>
      </section>
    </div>
  `;
  bindStatusBoardLinks();
}

function statusOverview(objects) {
  return {
    total: objects.length,
    stable: objects.filter((object) => yesish(object.stable)).length,
    candidate: objects.filter((object) => yesish(object.candidate)).length,
    withImplementation: objects.filter((object) => list(object.implementation_paths).length > 0).length
  };
}

function renderStatusRow(object) {
  return `
    <tr>
      <td><button class="table-object" type="button" data-node="${escapeAttr(objectNodeID(object))}">${escapeHTML(object.label)}</button></td>
      <td>${escapeHTML(object.human_state || object.layer || t("fallback.undeclared"))}</td>
      <td>${renderFlag(object.stable)}</td>
      <td>${renderFlag(object.candidate)}</td>
      <td>${renderCommandCell(object.next_label, object.next_command)}</td>
      <td>${escapeHTML(object.notes || t("fallback.none"))}</td>
    </tr>
  `;
}

function renderLifecycleCard(object) {
  const view = lifecycleView(object);
  return `
    <article class="lifecycle-card">
      <div class="lifecycle-head">
        <button class="card-object" type="button" data-node="${escapeAttr(objectNodeID(object))}">${escapeHTML(object.label)}</button>
        <span>${escapeHTML(object.human_state || object.layer || t("fallback.statusUnknown"))}</span>
      </div>
      ${renderLifecycleTrack(view, t("statusBoard.lifecycleAria", { label: object.label }))}
      <div class="progress-line ${view.complete ? "complete" : ""}"><span style="width: ${view.progress}%"></span></div>
      ${renderNextRoundEntry(view, object)}
      <p>${escapeHTML(t("fallback.nextStep", { value: object.next_command || t("fallback.undeclared") }))}</p>
    </article>
  `;
}

function lifecycleView(object, nextCommandOverride) {
  const command = String(nextCommandOverride || object.next_command || "").trim();
  const complete = isNextRoundEntry(object, command);
  const steps = lifecycleRoundSteps(object, command);
  let currentIndex = steps.findIndex((step) => step.command === command);
  if (complete) {
    currentIndex = steps.length;
  } else if (currentIndex < 0 && command) {
    steps.push(lifecycleStep(command));
    currentIndex = steps.length - 1;
  } else if (currentIndex < 0) {
    currentIndex = 0;
  }
  const progress = complete ? 100 : steps.length > 1 ? Math.round((currentIndex / (steps.length - 1)) * 100) : 0;
  return {
    steps,
    currentCommand: complete ? "" : command,
    currentIndex,
    progress,
    complete,
    nextRoundEntry: complete ? lifecycleStep(command) : null
  };
}

function lifecycleRoundSteps(object, command) {
  if (object.kind === "scenario") return scenarioRoundSteps(object, command);
  return unitRoundSteps(object, command);
}

function unitRoundSteps(object, command) {
  if (command === "unit_stable_verify") {
    return [lifecycleStep("unit_stable_verify"), lifecycleStep("unit_fork")];
  }
  if (isNextRoundEntry(object, command)) {
    return [
      lifecycleStep("unit_check"),
      lifecycleStep("unit_plan"),
      lifecycleStep("unit_impl"),
      lifecycleStep("unit_verify"),
      lifecycleStep("unit_promote")
    ];
  }
  if (command === "unit_init") {
    return [
      lifecycleStep("unit_init"),
      lifecycleStep("unit_new"),
      lifecycleStep("unit_check"),
      lifecycleStep("unit_plan"),
      lifecycleStep("unit_impl"),
      lifecycleStep("unit_verify"),
      lifecycleStep("unit_promote")
    ];
  }
  const startCommand = command === "unit_new" || !yesish(object.stable) ? "unit_new" : "unit_fork";
  return [
    lifecycleStep(startCommand),
    lifecycleStep("unit_check"),
    lifecycleStep("unit_plan"),
    lifecycleStep("unit_impl"),
    lifecycleStep("unit_verify"),
    lifecycleStep("unit_promote")
  ];
}

function scenarioRoundSteps(object, command) {
  if (command === "scenario_stable_verify") {
    return [lifecycleStep("scenario_stable_verify"), lifecycleStep("scenario_fork")];
  }
  if (isNextRoundEntry(object, command)) {
    return [
      lifecycleStep("scenario_check"),
      lifecycleStep("scenario_verify"),
      lifecycleStep("scenario_promote")
    ];
  }
  const startCommand = command === "scenario_new" || !yesish(object.stable) ? "scenario_new" : "scenario_fork";
  return [
    lifecycleStep(startCommand),
    lifecycleStep("scenario_check"),
    lifecycleStep("scenario_verify"),
    lifecycleStep("scenario_promote")
  ];
}

function isNextRoundEntry(object, command) {
  if (!yesish(object.stable) || yesish(object.candidate) || object.layer !== "stable") return false;
  return (object.kind === "unit" && command === "unit_fork") || (object.kind === "scenario" && command === "scenario_fork");
}

function renderLifecycleTrack(view, ariaLabel) {
  return `
    <div class="lifecycle-track" aria-label="${escapeAttr(ariaLabel)}">
      ${view.steps.map((step, index) => {
        const stateClass = view.complete || index < view.currentIndex ? "done" : index === view.currentIndex ? "current" : "future";
        return `
          <span class="lifecycle-step ${stateClass}" title="${escapeAttr(step.command + " · " + step.label)}">
            <code>${escapeHTML(step.command)}</code>
          </span>
        `;
      }).join("")}
    </div>
  `;
}

function renderNextRoundEntry(view, object) {
  if (!view.nextRoundEntry) return "";
  const intentClass = nextIntentClass(object);
  const label = nextRoundEntryLabel(object);
  const title = nextRoundEntryTitle(view.nextRoundEntry, object);
  return `
    <div class="next-round-entry ${escapeAttr(intentClass)}">
      <span>${escapeHTML(label)}</span>
      <span class="lifecycle-step current" title="${escapeAttr(title)}">
        <code>${escapeHTML(view.nextRoundEntry.command)}</code>
      </span>
    </div>
  `;
}

function nextRoundEntryLabel(object) {
  return nextIntent(object) === "repair" ? t("statusBoard.nextRepairEntry") : t("statusBoard.nextRoundEntry");
}

function nextRoundEntryTitle(step, object) {
  const intent = nextIntent(object);
  const label = intent ? todoIntentLabel(intent) : step.label;
  return `${step.command} · ${label}`;
}

function lifecycleStep(command) {
  return {
    command,
    short: t(`lifecycleShort.${command}`),
    label: t(`lifecycle.${command}`)
  };
}

function renderCommandCell(label, command) {
  if (!label && !command) return t("fallback.none");
  return `
    <div class="command-cell">
      <span>${escapeHTML(label || command)}</span>
      ${command ? `<code>${escapeHTML(command)}</code>` : ""}
    </div>
  `;
}

function renderFlag(value) {
  const active = yesish(value);
  return `<span class="flag ${active ? "flag-yes" : "flag-no"}">${escapeHTML(value || "no")}</span>`;
}

function renderSourceButton(path, label) {
  if (!path) return "";
  return `<button class="source-link" type="button" data-source="${escapeAttr(path)}">${escapeHTML(label)}</button>`;
}

function yesish(value) {
  return String(value || "").toLowerCase() === "yes";
}

function bindStatusBoardLinks() {
  graphView.querySelectorAll("[data-source]").forEach((button) => {
    button.addEventListener("click", (event) => {
      event.preventDefault();
      openSource(button.dataset.source);
    });
  });
  graphView.querySelectorAll("[data-node]").forEach((button) => {
    button.addEventListener("click", () => focusNode(button.dataset.node));
  });
}

function renderTodoNav() {
  const items = todoItems();
  const sections = todoTypeOrder()
    .map((type) => ({ type, items: items.filter((item) => item.type === type) }))
    .filter((section) => section.items.length > 0);
  if (sections.length === 0) {
    const empty = document.createElement("div");
    empty.className = "nav-empty";
    empty.textContent = t("todo.empty");
    navPanel.appendChild(empty);
    return;
  }
  if (!sections.some((section) => section.type === activeTodoNavGroup)) {
    activeTodoNavGroup = (sections[0] || {}).type || "other";
  }
  sections.forEach((section) => renderTodoNavSection(section.type, section.items));
}

function renderTodoNavSection(type, items) {
  const expanded = type === activeTodoNavGroup;
  const section = document.createElement("section");
  section.className = expanded ? "nav-section expanded" : "nav-section";

  const header = document.createElement("button");
  header.className = "nav-section-title";
  header.type = "button";
  header.setAttribute("aria-expanded", String(expanded));
  header.innerHTML = `<span>${escapeHTML(todoTypeLabel(type))}</span><em>${items.length}</em>`;
  header.addEventListener("click", () => {
    activeTodoNavGroup = type;
    renderNav();
  });
  section.appendChild(header);

  if (expanded) {
    items.forEach((item) => {
      const button = document.createElement("button");
      button.className = item.id === selectedNodeID ? "nav-item active" : "nav-item";
      button.type = "button";
      button.innerHTML = `<strong>${escapeHTML(item.objectLabel)}</strong><span>${escapeHTML(item.commandText)}</span>`;
      button.addEventListener("click", () => focusTodoItem(item.id));
      section.appendChild(button);
    });
  }
  navPanel.appendChild(section);
}

function renderTodoBoard() {
  const items = todoItems();
  if (items.length === 0) {
    graphView.innerHTML = `
      <section class="todo-empty-state">
        <h3>${escapeHTML(t("todo.emptyDetailTitle"))}</h3>
        <p>${escapeHTML(t("todo.emptyDetail"))}</p>
      </section>
    `;
    return;
  }
  graphView.innerHTML = `
    <section class="todo-board">
      <div class="todo-board-heading">
        <div>
          <h3>${escapeHTML(t("todo.boardHeading"))}</h3>
          <p>${escapeHTML(t("todo.boardDescription"))}</p>
        </div>
        ${renderSourceButton(snapshot.project.status_file, t("statusBoard.sourceLabel"))}
      </div>
      <div class="todo-card-grid">
        ${items.map(renderTodoCard).join("")}
      </div>
    </section>
  `;
  bindTodoBoardLinks();
}

function renderTodoCard(item) {
  const view = lifecycleView(item.object, item.nextCommand);
  return `
    <article class="todo-card ${item.id === selectedNodeID ? "active" : ""} ${escapeAttr(nextIntentClass(item.object))}" data-todo-card="${escapeAttr(item.id)}">
      <div class="todo-card-head">
        <button class="card-object" type="button" data-todo="${escapeAttr(item.id)}">${escapeHTML(item.objectLabel)}</button>
        <span class="todo-type ${escapeAttr(nextIntentClass(item.object))}">${escapeHTML(todoTypeLabel(item.type))}</span>
      </div>
      <div class="todo-command-row">
        <span>${escapeHTML(t("todo.command"))}</span>
        <div class="todo-command-actions">
          <button class="todo-copy-command" type="button" data-copy-next-command="${escapeAttr(item.commandText)}" title="${escapeAttr(t("review.copyNextCommand"))}">
            <code>${escapeHTML(item.commandText)}</code>
          </button>
          ${renderAdvanceCommandButton(item, "todo-copy-command advance-entry")}
        </div>
      </div>
      ${renderTodoIntentPill(item)}
      ${renderLifecycleTrack(view, t("statusBoard.lifecycleAria", { label: item.objectLabel }))}
      <div class="progress-line ${view.complete ? "complete" : ""}"><span style="width: ${view.progress}%"></span></div>
      ${renderNextRoundEntry(view, item.object)}
      <p>${escapeHTML(item.object.notes || t("fallback.none"))}</p>
    </article>
  `;
}

function renderTodoIntentPill(item) {
  const intent = nextIntent(item.object);
  if (!intent) return "";
  return `
    <div class="todo-intent ${escapeAttr(nextIntentClass(item.object))}">
      <span>${escapeHTML(t("todo.intent"))}</span>
      <strong>${escapeHTML(todoIntentLabel(intent))}</strong>
    </div>
  `;
}

function bindTodoBoardLinks() {
  graphView.querySelectorAll("[data-todo-card]").forEach((card) => {
    card.addEventListener("click", () => focusTodoItem(card.dataset.todoCard));
  });
  graphView.querySelectorAll("[data-todo]").forEach((button) => {
    button.addEventListener("click", (event) => {
      event.stopPropagation();
      focusTodoItem(button.dataset.todo);
    });
  });
  graphView.querySelectorAll("[data-source]").forEach((button) => {
    button.addEventListener("click", (event) => {
      event.preventDefault();
      event.stopPropagation();
      openSource(button.dataset.source);
    });
  });
  bindCopyCommandButtons(graphView);
}

function focusTodoItem(itemID) {
  const item = todoItemByID(itemID);
  if (item) activeTodoNavGroup = item.type;
  selectedNodeID = itemID;
  renderNav();
  renderGraph();
  renderDetailForNode(itemID);
}

function todoItems() {
  return list(snapshot.objects)
    .filter((object) => (object.kind === "unit" || object.kind === "scenario") && String(object.next_command || "").trim())
    .map((object) => {
      const nextCommand = String(object.next_command || "").trim();
      const type = todoTypeForObject(object, nextCommand);
      const sources = todoSourcesForObject(object, nextCommand);
      return {
        id: `todo:${object.kind}:${object.id}`,
        type,
        object,
        objectLabel: object.label || object.id || t("fallback.undeclared"),
        nextCommand,
        commandText: `${nextCommand}:${object.id}`,
        advanceCommandText: advanceEntryCommandForObject(object, nextCommand),
        sources,
        primarySources: sources.filter((source) => source.group !== "references"),
        referenceSources: sources.filter((source) => source.group === "references"),
        implementationPaths: list(object.implementation_paths).map((ref) => ref.path).filter(Boolean)
      };
    })
    .sort(compareTodoItems);
}

function todoItemByID(itemID) {
  return todoItems().find((item) => item.id === itemID) || null;
}

function todoItemForSource(path) {
  if (currentView !== "todo") return null;
  const selected = todoItemByID(selectedNodeID);
  if (selected && list(selected.sources).some((source) => source.path === path)) return selected;
  return todoItems().find((item) => list(item.sources).some((source) => source.path === path)) || null;
}

function todoTypeForCommand(command) {
  if (command === "unit_stable_verify" || command === "scenario_stable_verify") return "stableVerify";
  if (command === "unit_check" || command === "scenario_check") return "designCheck";
  if (command === "unit_plan") return "plan";
  if (command === "unit_impl") return "implementation";
  if (command === "unit_verify" || command === "scenario_verify") return "verify";
  if (command === "unit_promote" || command === "scenario_promote") return "promote";
  if (command === "unit_fork" || command === "scenario_fork") return "fork";
  if (command === "unit_init" || command === "unit_new" || command === "scenario_new") return "new";
  return "other";
}

function todoTypeForObject(object, command) {
  if (command === "unit_fork" && nextIntent(object) === "repair") return "repairFork";
  return todoTypeForCommand(command);
}

function todoTypeOrder() {
  return ["stableVerify", "repairFork", "designCheck", "plan", "implementation", "verify", "promote", "fork", "new", "other"];
}

function compareTodoItems(left, right) {
  return todoTypeOrder().indexOf(left.type) - todoTypeOrder().indexOf(right.type)
    || String(left.objectLabel || "").localeCompare(String(right.objectLabel || ""))
    || String(left.nextCommand || "").localeCompare(String(right.nextCommand || ""));
}

function advanceEntryCommandForObject(object, nextCommand) {
  const kind = String(object && object.kind ? object.kind : "").trim();
  const objectID = String(object && object.id ? object.id : "").trim();
  const command = String(nextCommand || "").trim();
  if (!kind || !objectID || !command) return "";
  if (kind === "unit" && ["unit_check", "unit_plan", "unit_impl", "unit_verify", "unit_promote"].includes(command)) {
    return `unit_advance:${objectID}`;
  }
  if (kind === "scenario" && ["scenario_check", "scenario_verify", "scenario_promote"].includes(command)) {
    return `scenario_advance:${objectID}`;
  }
  return "";
}

function renderAdvanceCommandButton(item, className) {
  const command = String(item && item.advanceCommandText ? item.advanceCommandText : "").trim();
  if (!command) return "";
  return `
    <button class="${escapeAttr(className)}" type="button" data-copy-next-command="${escapeAttr(command)}" title="${escapeAttr(t("todo.copyAdvanceEntry"))}">
      <span>${escapeHTML(t("todo.advanceEntry"))}</span>
      <code>${escapeHTML(command)}</code>
    </button>
  `;
}

function todoTypeLabel(type) {
  return t(`todo.types.${type}`);
}

function nextIntent(object) {
  return String(object && object.next_intent ? object.next_intent : "").trim();
}

function nextIntentClass(object) {
  const intent = nextIntent(object);
  return intent ? `intent-${intent}` : "";
}

function todoIntentLabel(intent) {
  return t(`todo.intents.${intent}`);
}

function todoSourcesForObject(object, command) {
  const sources = [];
  const addSource = (ref, labelKey, group = "materials") => {
    if (!ref || !ref.path || !sourceExists(ref.path)) return;
    if (sources.some((source) => source.path === ref.path && source.group === group)) return;
    sources.push({
      ...ref,
      label: t(`todo.sourceLabels.${labelKey}`),
      group
    });
  };
  const activeTruth = uniqueSources(object.truth_paths).filter((ref) => !isAppendixPath(ref.path));
  const appendices = uniqueSources(object.truth_paths).filter((ref) => isAppendixReference(ref));
  const evidence = uniqueSources(object.truth_paths).filter((ref) => isEvidenceReference(ref));
  const ruleSources = ruleSourcesForObject(object);

  if (command === "unit_check" || command === "scenario_check") {
    activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
    appendices.forEach((ref) => addSource(ref, "appendix"));
    evidence.forEach((ref) => addSource(ref, "evidence", "references"));
    return sources;
  }

  if (command === "unit_stable_verify" || command === "scenario_stable_verify") {
    activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
    appendices.forEach((ref) => addSource(ref, "appendix"));
    ruleSources.forEach((ref) => addSource(ref, "rule"));
    addSource(processSource(object, "verifyResult", "stable"), "verifyResult");
    return sources;
  }

  if (command === "unit_plan") {
    activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
    addSource(processSource(object, "checkResult"), "checkResult");
    return sources;
  }

  if (command === "unit_impl") {
    addSource(processSource(object, "activePlan"), "activePlan");
    activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
    return sources;
  }

  if (command === "unit_fork" || command === "scenario_fork") {
    activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
    appendices.forEach((ref) => addSource(ref, "appendix"));
    addSource(processSource(object, "verifyResult", "stable"), "verifyResult");
    return sources;
  }

  activeTruth.forEach((ref) => addSource(ref, "activeTruth"));
  addSource(processSource(object, "activePlan"), "activePlan");
  addSource(processSource(object, "checkResult"), "checkResult");
  addSource(processSource(object, "verifyResult", object.layer), "verifyResult");
  return sources;
}

function processSource(object, kind, layer) {
  if (!object || !object.id) return null;
  if (kind === "activePlan") return { path: `docs/specs/_plans/active/${object.id}.md` };
  if (kind === "checkResult") return { path: `docs/specs/_check_result/${object.kind}/${object.id}.md` };
  if (kind === "verifyResult") return { path: `docs/specs/_verify_result/${layer || object.layer || "candidate"}/${object.kind}/${object.id}.md` };
  return null;
}

function ruleSourcesForObject(object) {
  return list(object.rule_refs).flatMap((ruleID) => {
    const rule = list(snapshot.objects).find((item) => item.kind === "rule" && item.id === ruleID);
    return rule ? uniqueSources(rule.truth_paths) : [];
  });
}

function sourceExists(path) {
  if (!isReadableOriginalPath(path)) return false;
  return list(snapshot.sources).some((source) => source.path === path)
    || list(snapshot.nodes).some((node) => node.source && node.source.path === path)
    || list(snapshot.objects).some((object) => list(object.truth_paths).some((ref) => ref.path === path));
}

function isAppendixPath(path) {
  return String(path || "").includes("/appendix/");
}

function renderReviewBoard() {
  graphView.innerHTML = "";
}

function focusReviewItem(itemID) {
  const item = reviewItemByID(itemID);
  if (item) activeReviewNavGroup = item.reviewType;
  selectedNodeID = itemID;
  renderNav();
  renderGraph();
  renderDetailForNode(itemID);
}

function reviewItems() {
  const items = [];
  const seen = new Set();
  const addItem = (item) => {
    if (!item || !item.path || !isReadableOriginalPath(item.path)) return;
    const key = `${item.reviewType}:${item.path}:${item.object ? item.object.id : item.objectLabel}`;
    if (seen.has(key)) return;
    seen.add(key);
    items.push({
      ...item,
      id: `spec:${item.reviewType}:${item.path}:${item.object ? item.object.id : item.objectLabel}`,
      fileLabel: fileName(item.path),
      source: item.source || { path: item.path },
      nextCommand: item.object ? item.object.next_command : "",
      stateLabel: t(`review.states.${item.reviewType}`),
      targetType: item.targetType || reviewTargetTypeForObject(item.object)
    });
  };

  list(snapshot.objects).forEach((object) => {
    const targetType = reviewTargetTypeForObject(object);
    if (!targetType) return;
    if ((object.kind === "unit" || object.kind === "scenario") && object.layer === "candidate") {
      uniqueSources(object.truth_paths).filter((source) => isPrimaryReviewSource(source, object, "candidate")).forEach((source) => {
        addItem({
          reviewType: "candidate",
          targetType,
          path: source.path,
          source,
          object,
          objectLabel: object.label || object.id || t("fallback.undeclared")
        });
      });
      return;
    }
    if ((object.kind === "unit" || object.kind === "scenario") && object.layer === "stable") {
      uniqueSources(object.truth_paths).filter((source) => isPrimaryReviewSource(source, object, "stable")).forEach((source) => {
        addItem({
          reviewType: "stable",
          targetType,
          path: source.path,
          source,
          object,
          objectLabel: object.label || object.id || t("fallback.undeclared")
        });
      });
      return;
    }
    if (object.kind !== "rule" || object.layer !== "stable") return;
    uniqueSources(object.truth_paths).filter((source) => isPrimaryReviewSource(source, object, "stable")).forEach((source) => {
      addItem({
        reviewType: "stableRule",
        targetType,
        path: source.path,
        source,
        object,
        objectLabel: object.label || object.id || t("fallback.undeclared")
      });
    });
  });

  return items.sort(compareReviewItems);
}

function reviewItemByID(itemID) {
  return reviewItems().find((item) => item.id === itemID) || null;
}

function reviewItemByPath(path) {
  if (currentView !== "spec") return null;
  return reviewItems().find((item) => item.path === path) || null;
}

function reviewTargetTypeForObject(object) {
  if (!object) return "";
  if (object.kind === "unit") return "capability";
  if (object.kind === "scenario") return "scenario";
  if (object.kind === "rule") return "rule";
  return "";
}

function reviewTypeOrder() {
  return ["candidate", "stable", "stableRule"];
}

function compareReviewItems(left, right) {
  return reviewTypeOrder().indexOf(left.reviewType) - reviewTypeOrder().indexOf(right.reviewType)
    || Number(Boolean(right.nextCommand)) - Number(Boolean(left.nextCommand))
    || String(left.objectLabel || "").localeCompare(String(right.objectLabel || ""))
    || String(left.path || "").localeCompare(String(right.path || ""));
}

function isPrimaryReviewSource(source, object, layer) {
  const path = String(source && source.path ? source.path : "");
  if (!path.includes(`/${layer}/`)) return false;
  if (isEvidenceReference(source)) return false;
  const name = fileName(path);
  const prefix = layer === "stable" ? "s" : "c";
  if (object.kind === "unit") return new RegExp(`^${prefix}_unit_[^/]+\\.md$`).test(name);
  if (object.kind === "scenario") return new RegExp(`^${prefix}_scenario_[^/]+\\.md$`).test(name);
  if (object.kind === "rule") return new RegExp(`^${prefix}_[gb]_rule_[^/]+\\.md$`).test(name);
  return false;
}

function isAppendixReference(source) {
  const path = String(source && source.path ? source.path : "");
  return path.includes("/appendix/") && !isEvidenceReference(source);
}

function isEvidenceReference(source) {
  const path = String(source && source.path ? source.path : "");
  return path.includes("/appendix/") && /_evidence\.md$/.test(fileName(path));
}

function isStableReference(source) {
  const path = String(source && source.path ? source.path : "");
  return path.includes("/stable/") || fileName(path).startsWith("s_");
}

function reviewTypeLabel(type) {
  return t(`review.types.${type}`);
}

function reviewTarget(type) {
  return t(`review.targets.${type}`);
}

function reviewFocusItems(type) {
  return String(t(`review.focus.${type}`))
    .split(/[、,]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function reviewNavSubtitle(item) {
  const state = item.stateLabel || reviewTypeLabel(item.reviewType);
  return `${state} · ${item.objectLabel}`;
}

function reviewNextCommandText(item) {
  const command = String(item && item.nextCommand ? item.nextCommand : "").trim();
  const objectID = String(item && item.object && item.object.id ? item.object.id : "").trim();
  if (!command || !objectID) return "";
  return `${command}:${objectID}`;
}

function renderReviewProgressHeader(path) {
  const item = currentView === "todo" ? todoItemForSource(path) : reviewItemByPath(path);
  if (!item || !item.object) return "";
  if (item.object.kind !== "unit" && item.object.kind !== "scenario") return "";
  const view = lifecycleView(item.object, item.nextCommand);
  const command = item.commandText || reviewNextCommandText(item);
  const advanceItem = {
    advanceCommandText: item.advanceCommandText || advanceEntryCommandForObject(item.object, item.nextCommand)
  };
  if (command) {
    return `
      <section class="review-progress-panel">
        <div class="review-progress-head">
          <h2>${escapeHTML(t("review.progressTitle"))}</h2>
          <div class="review-command-actions">
            <button class="review-next-command" type="button" data-copy-next-command="${escapeAttr(command)}" title="${escapeAttr(t("review.copyNextCommand"))}">
              <span>${escapeHTML(t("review.nextCommand"))}</span>
              <code>${escapeHTML(command)}</code>
            </button>
            ${renderAdvanceCommandButton(advanceItem, "review-next-command advance-entry")}
          </div>
        </div>
        ${renderLifecycleTrack(view, t("statusBoard.lifecycleAria", { label: item.objectLabel }))}
        <div class="progress-line ${view.complete ? "complete" : ""}"><span style="width: ${view.progress}%"></span></div>
        ${renderNextRoundEntry(view, item.object)}
      </section>
    `;
  }
  return `
    <section class="review-progress-panel">
      <div class="review-progress-head">
        <h2>${escapeHTML(t("review.progressTitle"))}</h2>
        <span class="review-next-empty">${escapeHTML(t("review.noNextCommand"))}</span>
      </div>
      ${renderLifecycleTrack(view, t("statusBoard.lifecycleAria", { label: item.objectLabel }))}
      <div class="progress-line ${view.complete ? "complete" : ""}"><span style="width: ${view.progress}%"></span></div>
      ${renderNextRoundEntry(view, item.object)}
    </section>
  `;
}

function bindReviewProgressHeader() {
  bindCopyCommandButtons(sourceRendered);
  bindCopyCommandButtons(detailPanel);
}

function bindCopyCommandButtons(root) {
  root.querySelectorAll("[data-copy-next-command]").forEach((button) => {
    if (button.dataset.copyBound === "true") return;
    button.dataset.copyBound = "true";
    button.addEventListener("click", async (event) => {
      event.preventDefault();
      event.stopPropagation();
      const command = button.dataset.copyNextCommand || "";
      const originalHTML = button.innerHTML;
      try {
        await navigator.clipboard.writeText(command);
        button.textContent = t("review.copied");
      } catch {
        button.textContent = t("review.copyFailed");
      }
      window.setTimeout(() => {
        button.innerHTML = originalHTML;
      }, 1200);
    });
  });
}

function reviewRelationSummary(item) {
  const parts = reviewRelationGroups(item)
    .map((group) => `${group.label} ${group.items.length}`);
  return parts.length > 0 ? parts.join(" · ") : t("review.relationEmpty");
}

function reviewRelationGroups(item) {
  const object = item ? item.object : null;
  if (!object) return [];
  const groups = [];
  const implementation = list(object.implementation_paths).map((ref) => ref.path).filter(Boolean);
  if (implementation.length > 0) groups.push({ label: t("review.relation.implementation"), items: implementation, linkable: false });
  const ruleFiles = ruleSourcesForObject(object).map((ref) => ref.path).filter(Boolean);
  if (ruleFiles.length > 0) {
    groups.push({ label: t("review.relation.ruleFile"), items: ruleFiles, linkable: true });
  } else {
    const rules = list(object.rule_refs).filter(Boolean);
    if (rules.length > 0) groups.push({ label: t("review.relation.rule"), items: rules, linkable: false });
  }
  const bound = list(object.bound_objects).filter(Boolean);
  if (bound.length > 0) groups.push({ label: t("review.relation.bound"), items: bound, linkable: false });
  const appendix = uniqueSources(object.truth_paths)
    .filter((ref) => isAppendixReference(ref))
    .map((ref) => ref.path);
  if (appendix.length > 0) groups.push({ label: t("review.relation.appendix"), items: appendix, linkable: true });
  const evidence = uniqueSources(object.truth_paths)
    .filter((ref) => isEvidenceReference(ref))
    .map((ref) => ref.path);
  if (evidence.length > 0) groups.push({ label: t("review.relation.evidence"), items: evidence, linkable: true });
  const stable = uniqueSources(object.truth_paths)
    .filter((ref) => isStableReference(ref))
    .map((ref) => ref.path);
  if (stable.length > 0) groups.push({ label: t("review.relation.stable"), items: stable, linkable: true });
  if (snapshot.project.mapping_file) {
    groups.push({ label: t("review.relation.mapping"), items: [snapshot.project.mapping_file], linkable: true });
  }
  if (snapshot.project.rule_baseline_file) {
    groups.push({ label: t("review.relation.system"), items: [snapshot.project.rule_baseline_file], linkable: true });
  }
  return groups;
}

function renderReviewDetail(item) {
  if (!item) {
    detailPanel.innerHTML = `<h2>${escapeHTML(t("fallback.noObject"))}</h2>`;
    updateTruthTab([]);
    return;
  }
  detailPanel.innerHTML = `
    <h2>${escapeHTML(item.fileLabel)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("review.fileType"))}</dt><dd>${escapeHTML(reviewTypeLabel(item.reviewType))}</dd>
      <dt>${escapeHTML(t("review.object"))}</dt><dd>${escapeHTML(item.objectLabel)}</dd>
      <dt>${escapeHTML(t("inspector.fields.status"))}</dt><dd>${escapeHTML(item.stateLabel || reviewTypeLabel(item.reviewType))}</dd>
      <dt>${escapeHTML(t("inspector.fields.file"))}</dt><dd>${escapeHTML(item.path)}</dd>
    </dl>
    <section class="review-detail-section">
      <h2>${escapeHTML(t("review.reviewTarget"))}</h2>
      <p>${escapeHTML(reviewTarget(item.reviewType))}</p>
    </section>
    <section class="review-detail-section">
      <h2>${escapeHTML(t("review.readingFocus"))}</h2>
      <ul class="review-focus-points">
        ${reviewFocusItems(item.reviewType).map((focus) => `<li>${escapeHTML(focus)}</li>`).join("")}
      </ul>
    </section>
    <section class="review-detail-section">
      <h2>${escapeHTML(t("review.relationships"))}</h2>
      ${renderReviewRelationGroups(item)}
    </section>
    <section class="review-detail-section">
      ${renderSourceButton(item.path, t("review.openSource"))}
    </section>
  `;
  bindInspectorLinks();
  updateTruthTab([item.source], item.id, { activate: true });
}

function renderReviewEmptyDetail() {
  detailPanel.innerHTML = `
    <section class="review-empty-state">
      <h2>${escapeHTML(t("review.emptyDetailTitle"))}</h2>
      <p>${escapeHTML(t("review.emptyDetail"))}</p>
    </section>
  `;
  updateTruthTab([], "review-empty");
}

function renderReviewRelationGroups(item) {
  const groups = reviewRelationGroups(item);
  if (groups.length === 0) return `<p class="empty-copy">${escapeHTML(t("review.relationEmpty"))}</p>`;
  return groups.map((group) => {
    const chips = group.items.map((value) => {
      if (group.linkable) {
        return `<button class="chip" type="button" data-source="${escapeAttr(value)}">${escapeHTML(value)}</button>`;
      }
      return `<span class="chip">${escapeHTML(value)}</span>`;
    }).join("");
    return `<h3 class="review-relation-title">${escapeHTML(group.label)}</h3><div class="chips">${chips}</div>`;
  }).join("");
}

function fileName(path) {
  return String(path || "").split("/").pop() || String(path || "");
}

function compactLabel(node) {
  const label = String(node.label || "");
  if (node.kind === "project_path" || node.kind === "project_area") {
    return label
      .replace(/^docs\/specs\/units\/candidate\/appendix\//, "appendix/")
      .replace(/^docs\/specs\/units\/candidate\//, "units/candidate/")
      .replace(/^docs\/specs\/units\/stable\//, "units/stable/")
      .replace(/^docs\/specs\/rules\/candidate\//, "rule/candidate/")
      .replace(/^docs\/specs\/rules\/stable\//, "rule/stable/")
      .replace("/**", "\n/**");
  }
  if (node.kind === "truth_file") {
    return compactTruthFileLabel(label);
  }
  if (node.kind === "implementation_path") {
    return label.replace("/**", "\n/**");
  }
  return label;
}

function compactTruthFileLabel(label) {
  const base = String(label || "").replace(/\.md$/, "");
  const unitMatch = base.match(/^([cs])_unit_(.+)$/);
  if (unitMatch) return `${unitMatch[2].replace(/_/g, " ")} (${truthLayerName(unitMatch[1])})`;
  const sharedMatch = base.match(/^([cs])_[gb]_rule_(.+)$/);
  if (sharedMatch) return `rule ${sharedMatch[2].replace(/_/g, " ")} (${truthLayerName(sharedMatch[1])})`;
  return base.replace(/_/g, " ");
}

function truthLayerName(prefix) {
  return prefix === "s" ? "stable" : "candidate";
}

function edgeLabel(kind) {
  if (kind === "described_by") return "Spec";
  if (kind === "owns_path") return "Path";
  if (kind === "uses_rule") return "Uses";
  if (kind === "bound_to") return "Bound";
  if (kind === "contains") return "Contains";
  if (kind === "maps_to") return "Owner";
  if (kind === "declares") return "Declares";
  if (kind === "tracks_state") return "State";
  if (kind === "constrains") return "Constrains";
  return kind;
}

function nodeSize(ele) {
  const group = ele.data("group");
  if (group === "root") return 48;
  if (group === "unit" || group === "scenario") return 42;
  if (group === "rule") return 40;
  if (group === "truth") return 34;
  return 36;
}

function edgeWidth(ele) {
  const kind = ele.data("kind");
  if (kind === "uses_rule" || kind === "bound_to" || kind === "maps_to") return 2;
  return 1.5;
}

function average(values) {
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function byLabel(left, right) {
  return String(left.label || "").localeCompare(String(right.label || ""));
}

function renderDetail(object) {
  if (!object) {
    detailPanel.innerHTML = `<h2>${escapeHTML(t("fallback.noObject"))}</h2>`;
    updateTruthTab([]);
    return;
  }
  const truthRefs = truthRefsForObject(object);
  detailPanel.innerHTML = `
    <h2>${escapeHTML(object.label)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("inspector.fields.type"))}</dt><dd>${escapeHTML(object.kind)}</dd>
      <dt>${escapeHTML(t("inspector.fields.status"))}</dt><dd>${escapeHTML(object.human_state || t("fallback.undeclared"))}</dd>
      <dt>${escapeHTML(t("inspector.fields.version"))}</dt><dd>${escapeHTML(object.version || t("fallback.undeclared"))}</dd>
      <dt>${escapeHTML(t("inspector.fields.next"))}</dt><dd>${escapeHTML(object.next_label || object.next_command || t("fallback.none"))}</dd>
      <dt>${escapeHTML(t("inspector.fields.responsibility"))}</dt><dd>${escapeHTML(object.responsibility || t("fallback.undeclared"))}</dd>
      <dt>${escapeHTML(t("inspector.fields.notes"))}</dt><dd>${escapeHTML(object.notes || t("fallback.none"))}</dd>
    </dl>
    ${renderChipGroup(t("inspector.groups.truth"), object.truth_paths, true)}
    ${renderImplementationPathGroup(t("inspector.groups.implementation"), object.implementation_paths)}
    ${renderTextChips(t("inspector.groups.rule"), object.rule_refs)}
    ${renderTextChips(t("inspector.groups.bound"), object.bound_objects)}
  `;
  bindInspectorLinks();
  updateTruthTab(truthRefs, objectNodeID(object));
}

function renderDetailForNode(nodeID) {
  if (currentView === "todo") {
    const item = todoItemByID(nodeID);
    if (item) {
      renderTodoDetail(item);
      return;
    }
    renderTodoEmptyDetail();
    return;
  }
  if (currentView === "spec") {
    const item = reviewItemByID(nodeID);
    if (item) {
      renderReviewDetail(item);
      return;
    }
    renderReviewEmptyDetail();
    return;
  }
  const object = objectFromNode(nodeID);
  if (object) {
    renderDetail(object);
    return;
  }
  const graph = graphForCurrentView();
  const node = graph.nodes.find((item) => item.id === nodeID);
  if (!node) {
    detailPanel.innerHTML = `<h2>${escapeHTML(t("fallback.noObject"))}</h2>`;
    updateTruthTab([], nodeID);
    return;
  }
  if (currentView === "project" && node.kind === "project_area") {
    renderProjectAreaDetail(node, graph);
    return;
  }
  if (currentView === "project" && node.kind === "project_root") {
    renderProjectRootDetail(node, graph);
    return;
  }
  const truthRefs = truthRefsForNode(node);
  const outgoing = graph.edges.filter((edge) => edge.from === nodeID);
  const incoming = graph.edges.filter((edge) => edge.to === nodeID);
  const connected = outgoing.concat(incoming).map((edge) => edge.from === nodeID ? edge.to : edge.from);
  const connectedNodes = connected
    .map((id) => graph.nodes.find((item) => item.id === id))
    .filter(Boolean);
  detailPanel.innerHTML = `
    <h2>${escapeHTML(node.label)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("inspector.fields.type"))}</dt><dd>${escapeHTML(labelForKind(node.kind))}</dd>
      ${node.source && node.source.path ? `<dt>${escapeHTML(t("inspector.fields.file"))}</dt><dd>${escapeHTML(node.source.path)}</dd>` : ""}
      <dt>${escapeHTML(t("inspector.fields.connections"))}</dt><dd>${incoming.length + outgoing.length}</dd>
    </dl>
    ${renderChipGroup(t("inspector.groups.truth"), truthRefs, true)}
    ${renderNodeList(t("inspector.groups.connected"), connectedNodes)}
  `;
  bindInspectorLinks();
  updateTruthTab(truthRefs, nodeID);
}

function renderProjectAreaDetail(node, graph) {
  const ownerNodes = graph.edges
    .filter((edge) => edge.from === node.id && edge.kind === "maps_to")
    .map((edge) => graph.nodes.find((item) => item.id === edge.to))
    .filter(Boolean)
    .sort(byLabel);
  const truthRefs = uniqueSources(list(node.raw_paths).map(sourceForImplementationRef));
  detailPanel.innerHTML = `
    <h2>${escapeHTML(node.aggregate_path_label || node.label)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("inspector.fields.type"))}</dt><dd>${escapeHTML(labelForKind(node.kind))}</dd>
      <dt>${escapeHTML(t("inspector.fields.connections"))}</dt><dd>${escapeHTML(ownerNodes.length)}</dd>
      <dt>${escapeHTML(t("inspector.fields.paths"))}</dt><dd>${escapeHTML(t("counts.paths", { count: list(node.raw_paths).length }))}</dd>
    </dl>
    ${renderImplementationPathGroup(t("inspector.groups.implementation"), node.raw_paths)}
    ${renderNodeList(t("inspector.groups.connected"), ownerNodes)}
  `;
  bindInspectorLinks();
  updateTruthTab(truthRefs, node.id);
}

function renderProjectRootDetail(node, graph) {
  const areaNodes = graph.edges
    .filter((edge) => edge.from === node.id && edge.kind === "contains")
    .map((edge) => graph.nodes.find((item) => item.id === edge.to))
    .filter(Boolean)
    .sort(byLabel);
  const truthRefs = uniqueSources(list(node.raw_paths).map(sourceForImplementationRef));
  detailPanel.innerHTML = `
    <h2>${escapeHTML(node.label)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("inspector.fields.type"))}</dt><dd>${escapeHTML(labelForKind(node.kind))}</dd>
      <dt>${escapeHTML(t("inspector.fields.connections"))}</dt><dd>${escapeHTML(areaNodes.length)}</dd>
      <dt>${escapeHTML(t("inspector.fields.paths"))}</dt><dd>${escapeHTML(t("counts.paths", { count: list(node.raw_paths).length }))}</dd>
    </dl>
    ${renderNodeList(t("views.project.groups.areas"), areaNodes)}
  `;
  bindInspectorLinks();
  updateTruthTab(truthRefs, node.id);
}

function renderTodoDetail(item) {
  const view = lifecycleView(item.object, item.nextCommand);
  detailPanel.innerHTML = `
    <h2>${escapeHTML(item.objectLabel)}</h2>
    <dl class="detail-grid">
      <dt>${escapeHTML(t("todo.actionType"))}</dt><dd>${escapeHTML(todoTypeLabel(item.type))}</dd>
      <dt>${escapeHTML(t("todo.command"))}</dt><dd><code>${escapeHTML(item.commandText)}</code></dd>
      ${renderTodoIntentDetailRows(item)}
      <dt>${escapeHTML(t("inspector.fields.status"))}</dt><dd>${escapeHTML(item.object.human_state || item.object.layer || t("fallback.undeclared"))}</dd>
      <dt>${escapeHTML(t("todo.notes"))}</dt><dd>${escapeHTML(item.object.notes || t("fallback.none"))}</dd>
    </dl>
    <section class="todo-detail-section">
      <h2>${escapeHTML(t("review.progressTitle"))}</h2>
      ${renderLifecycleTrack(view, t("statusBoard.lifecycleAria", { label: item.objectLabel }))}
      <div class="progress-line ${view.complete ? "complete" : ""}"><span style="width: ${view.progress}%"></span></div>
      ${renderNextRoundEntry(view, item.object)}
      <div class="review-command-actions">
        <button class="review-next-command" type="button" data-copy-next-command="${escapeAttr(item.commandText)}" title="${escapeAttr(t("review.copyNextCommand"))}">
          <span>${escapeHTML(t("review.nextCommand"))}</span>
          <code>${escapeHTML(item.commandText)}</code>
        </button>
        ${renderAdvanceCommandButton(item, "review-next-command advance-entry")}
      </div>
    </section>
    <section class="todo-detail-section">
      <h2>${escapeHTML(t("todo.materials"))}</h2>
      ${renderTodoSourceList(item.primarySources)}
    </section>
    ${item.referenceSources.length > 0 ? `
      <section class="todo-detail-section">
        <h2>${escapeHTML(t("todo.references"))}</h2>
        ${renderTodoSourceList(item.referenceSources)}
      </section>
    ` : ""}
    ${item.implementationPaths.length > 0 ? `
      <section class="todo-detail-section">
        <h2>${escapeHTML(t("todo.implementation"))}</h2>
        <div class="chips">${item.implementationPaths.map((path) => `<span class="chip">${escapeHTML(path)}</span>`).join("")}</div>
      </section>
    ` : ""}
  `;
  bindInspectorLinks();
  bindReviewProgressHeader();
  updateTruthTab(item.sources, item.id);
}

function renderTodoIntentDetailRows(item) {
  const intent = nextIntent(item.object);
  if (!intent) return "";
  return `<dt>${escapeHTML(t("todo.intent"))}</dt><dd>${escapeHTML(todoIntentLabel(intent))}</dd>`;
}

function renderTodoEmptyDetail() {
  detailPanel.innerHTML = `
    <section class="review-empty-state">
      <h2>${escapeHTML(t("todo.emptyDetailTitle"))}</h2>
      <p>${escapeHTML(t("todo.emptyDetail"))}</p>
    </section>
  `;
  updateTruthTab([], "todo-empty");
}

function renderTodoSourceList(sources) {
  if (!sources || sources.length === 0) return `<p class="empty-copy">${escapeHTML(t("todo.noMaterials"))}</p>`;
  return `
    <div class="todo-source-list">
      ${sources.map((source) => `
        <button class="todo-source" type="button" data-source="${escapeAttr(source.path)}">
          <span>${escapeHTML(source.label || t("todo.openMaterial"))}</span>
          <code>${escapeHTML(source.path)}</code>
        </button>
      `).join("")}
    </div>
  `;
}

function bindInspectorLinks() {
  detailPanel.querySelectorAll("[data-source]").forEach((link) => {
    link.addEventListener("click", (event) => {
      event.preventDefault();
      const line = Number(link.dataset.sourceLine || 0);
      openSource(link.dataset.source, line > 0 ? { line } : {});
    });
  });
  detailPanel.querySelectorAll("[data-node]").forEach((button) => {
    button.addEventListener("click", () => focusNode(button.dataset.node));
  });
}

function truthRefsForObject(object) {
  return uniqueSources((object.truth_paths || []).concat(object.sources || []));
}

function truthRefsForNode(node) {
  if (node && list(node.raw_paths).length > 0) {
    return uniqueSources(list(node.raw_paths).map(sourceForImplementationRef));
  }
  if (!node || !node.source || !isReadableOriginalPath(node.source.path)) return [];
  return [node.source];
}

function uniqueSources(sources) {
  const seen = new Set();
  return list(sources).filter(Boolean).filter((source) => {
    if (!source.path || seen.has(source.path)) return false;
    seen.add(source.path);
    return true;
  });
}

function updateTruthTab(truthRefs, ownerID, options = {}) {
  const refs = uniqueSources(truthRefs).filter((ref) => isReadableOriginalPath(ref.path));
  const hasTruth = refs.length > 0;
  const activateTruth = options.activate === true;
  truthTab.classList.toggle("hidden", !hasTruth);
  if (!hasTruth) {
    activeTruthOwnerID = null;
    sourcePath.textContent = "";
    sourceContent.textContent = t("source.emptyRaw");
    sourceRendered.textContent = t("source.emptyRendered");
    activeSourceHeadings = [];
    renderDocGuide([]);
    setInspectorTab("info");
    return;
  }
  if (activeTruthOwnerID !== ownerID || !refs.some((ref) => ref.path === sourcePath.textContent)) {
    activeTruthOwnerID = ownerID;
    openSource(refs[0].path, { activate: activateTruth });
  }
  setInspectorTab(activateTruth || activeInspectorTab === "truth" ? "truth" : "info");
}

function setInspectorTab(tabName) {
  if (tabName === "truth" && truthTab.classList.contains("hidden")) {
    tabName = "info";
  }
  activeInspectorTab = tabName;
  infoTab.classList.toggle("active", tabName === "info");
  truthTab.classList.toggle("active", tabName === "truth");
  detailPanel.classList.toggle("hidden", tabName !== "info");
  truthPanel.classList.toggle("hidden", tabName !== "truth");
}

function setDocMode(mode) {
  activeDocMode = mode === "raw" ? "raw" : "rendered";
  document.querySelectorAll("[data-doc-mode]").forEach((button) => {
    button.classList.toggle("active", button.dataset.docMode === activeDocMode);
  });
  sourceRendered.classList.toggle("hidden", activeDocMode !== "rendered");
  sourceContent.classList.toggle("hidden", activeDocMode !== "raw");
}

function isReadableOriginalPath(path) {
  if (!path) return false;
  return path.startsWith("docs/specs/")
    || path.startsWith("docs/project_standards/")
    || path === "AGENTS.md"
    || path === "CLAUDE.md"
    || path === "GEMINI.md";
}

function startInspectorResize(event) {
  event.preventDefault();
  resizeBar.classList.add("dragging");
  const minWidth = 220;
  const onPointerMove = (moveEvent) => {
    const nextWidth = Math.round(window.innerWidth - moveEvent.clientX - 12);
    const clampedWidth = Math.max(minWidth, nextWidth);
    document.documentElement.style.setProperty("--inspector-width", `${clampedWidth}px`);
    if (cy) cy.resize();
  };
  const stopResize = () => {
    resizeBar.classList.remove("dragging");
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", stopResize);
    if (cy) cy.resize();
  };
  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", stopResize);
}

function renderNodeList(title, nodes) {
  if (!nodes || nodes.length === 0) return "";
  return `<h2>${title}</h2><div class="chips">${nodes.map((node) => `<button class="chip" type="button" data-node="${escapeAttr(node.id)}">${escapeHTML(node.label)}</button>`).join("")}</div>`;
}

function labelForKind(kind) {
  const translated = lookupTranslation(TRANSLATIONS[currentLanguage], `kind.${kind}`)
    ?? lookupTranslation(TRANSLATIONS["zh-CN"], `kind.${kind}`);
  if (translated) return translated;
  return kind;
}

function renderImplementationPathGroup(title, refs) {
  if (!refs || refs.length === 0) return "";
  const chips = refs.map((ref) => {
    const source = sourceForImplementationRef(ref);
    if (source && isReadableOriginalPath(source.path)) {
      return `<button class="chip" type="button" data-source="${escapeAttr(source.path)}" ${source.line ? `data-source-line="${escapeAttr(source.line)}"` : ""}>${escapeHTML(ref.path)}</button>`;
    }
    return `<span class="chip">${escapeHTML(ref.path)}</span>`;
  }).join("");
  return `<h2>${title}</h2><div class="chips">${chips}</div>`;
}

function sourceForImplementationRef(ref) {
  if (!ref) return null;
  const sourcePath = ref.label && isReadableOriginalPath(ref.label) ? ref.label : snapshot.project.mapping_file;
  if (!sourcePath) return null;
  return { path: sourcePath, line: ref.line || 0, label: ref.path || sourcePath };
}

function renderChipGroup(title, refs, linkable) {
  if (!refs || refs.length === 0) return "";
  const chips = refs.map((ref) => {
    if (linkable) {
      return `<button class="chip" type="button" data-source="${escapeAttr(ref.path)}">${escapeHTML(ref.path)}</button>`;
    }
    return `<span class="chip">${escapeHTML(ref.path)}</span>`;
  }).join("");
  return `<h2>${title}</h2><div class="chips">${chips}</div>`;
}

function renderTextChips(title, items) {
  if (!items || items.length === 0) return "";
  return `<h2>${title}</h2><div class="chips">${items.map((item) => `<span class="chip">${escapeHTML(item)}</span>`).join("")}</div>`;
}

function renderSources(sources) {
  const seen = new Set();
  return list(sources).filter(Boolean).filter((source) => {
    if (!source.path || seen.has(source.path)) return false;
    seen.add(source.path);
    return true;
  }).map((source) => `<a href="#" class="source-link" data-source="${escapeAttr(source.path)}">${escapeHTML(source.path)}${source.line ? ":" + source.line : ""}</a>`).join("");
}

async function openSource(path, options = {}) {
  const activate = options.activate !== false;
  const targetLine = Number(options.line || 0);
  const response = await fetch(`/api/source?path=${encodeURIComponent(path)}`);
  if (!response.ok) {
    const message = await response.text();
    sourcePath.textContent = path;
    sourceContent.textContent = message;
    sourceRendered.textContent = message;
    activeSourceHeadings = [];
    renderDocGuide([]);
    setDocMode(activeDocMode);
    if (activate) setInspectorTab("truth");
    return;
  }
  const source = await response.json();
  const renderedDoc = renderMarkdownDocument(source.content);
  sourcePath.textContent = source.path;
  sourceContent.textContent = source.content;
  sourceRendered.innerHTML = renderReviewProgressHeader(source.path) + renderedDoc.html;
  activeSourceHeadings = renderedDoc.headings;
  renderDocGuide(activeSourceHeadings);
  bindReviewProgressHeader();
  bindRenderedDocLinks(source.path);
  bindDocGuideLinks();
  renderMermaidBlocks();
  if (targetLine > 0) {
    setDocMode("raw");
    requestAnimationFrame(() => scrollRawSourceToLine(targetLine));
  } else {
    setDocMode(activeDocMode);
  }
  if (activate) setInspectorTab("truth");
}

function renderMarkdown(markdown) {
  return renderMarkdownDocument(markdown).html;
}

function renderMarkdownDocument(markdown) {
  const parsed = splitFrontmatter(String(markdown || "").replaceAll("\r\n", "\n"));
  const lines = parsed.body.split("\n");
  const html = [];
  const headings = [];
  let paragraph = [];
  let listType = "";
  let inCode = false;
  let codeLines = [];
  let codeLang = "";
  let tableLines = [];

  const flushParagraph = () => {
    if (paragraph.length === 0) return;
    html.push(`<p>${renderInline(paragraph.join(" "))}</p>`);
    paragraph = [];
  };
  const flushList = () => {
    if (!listType) return;
    html.push(`</${listType}>`);
    listType = "";
  };
  const flushTable = () => {
    if (tableLines.length === 0) return;
    html.push(renderTable(tableLines));
    tableLines = [];
  };
  const flushBlocks = () => {
    flushParagraph();
    flushList();
    flushTable();
  };

  if (parsed.frontmatter.length > 0) {
    html.push(renderFrontmatter(parsed.frontmatter));
  }

  lines.forEach((line, index) => {
    if (line.startsWith("```")) {
      if (inCode) {
        const code = codeLines.join("\n");
        if (codeLang === "mermaid" || codeLang === "mmd") {
          html.push(`<div class="mermaid">${escapeHTML(code)}</div>`);
        } else {
          html.push(`<pre><code>${escapeHTML(code)}</code></pre>`);
        }
        codeLines = [];
        codeLang = "";
        inCode = false;
      } else {
        flushBlocks();
        inCode = true;
        codeLang = line.slice(3).trim().split(/\s+/)[0].toLowerCase();
      }
      return;
    }
    if (inCode) {
      codeLines.push(line);
      return;
    }

    const trimmed = line.trim();
    if (trimmed === "") {
      flushBlocks();
      return;
    }
    if (isTableLine(trimmed)) {
      flushParagraph();
      flushList();
      tableLines.push(trimmed);
      return;
    }
    flushTable();

    const heading = /^(#{1,4})\s+(.+)$/.exec(trimmed);
    if (heading) {
      flushParagraph();
      flushList();
      const level = heading[1].length;
      const text = plainHeadingText(heading[2]);
      const id = `doc-heading-${headings.length + 1}`;
      headings.push({ id, level, text, line: parsed.bodyStartLine + index });
      html.push(`<h${level} id="${escapeAttr(id)}">${renderInline(heading[2])}</h${level}>`);
      return;
    }

    const unordered = /^[-*]\s+(.+)$/.exec(trimmed);
    if (unordered) {
      flushParagraph();
      if (listType && listType !== "ul") flushList();
      if (!listType) {
        listType = "ul";
        html.push("<ul>");
      }
      html.push(`<li>${renderInline(unordered[1])}</li>`);
      return;
    }

    const ordered = /^\d+\.\s+(.+)$/.exec(trimmed);
    if (ordered) {
      flushParagraph();
      if (listType && listType !== "ol") flushList();
      if (!listType) {
        listType = "ol";
        html.push("<ol>");
      }
      html.push(`<li>${renderInline(ordered[1])}</li>`);
      return;
    }

    if (trimmed.startsWith("> ")) {
      flushParagraph();
      flushList();
      html.push(`<blockquote>${renderInline(trimmed.slice(2))}</blockquote>`);
      return;
    }

    flushList();
    paragraph.push(trimmed);
  });

  if (inCode) {
    html.push(`<pre><code>${escapeHTML(codeLines.join("\n"))}</code></pre>`);
  }
  flushBlocks();
  return { html: html.join(""), headings };
}

function renderDocGuide(headings) {
  const items = list(headings);
  if (!docGuide) return;
  if (items.length === 0) {
    docGuide.classList.add("hidden");
    docGuide.innerHTML = "";
    truthPanel.classList.add("guide-closed");
    updateDocGuideToggle();
    return;
  }
  docGuide.classList.toggle("hidden", !docGuideOpen);
  truthPanel.classList.toggle("guide-closed", !docGuideOpen);
  docGuide.innerHTML = `
    <div class="doc-guide-title">${escapeHTML(t("source.guideTitle"))}</div>
    <div class="doc-guide-list">
      ${items.map((heading) => `
        <button class="doc-guide-item depth-${Math.min(Math.max(heading.level, 1), 4)}" type="button" data-heading-id="${escapeAttr(heading.id)}" data-heading-line="${escapeAttr(heading.line)}">
          ${escapeHTML(heading.text || t("source.noGuide"))}
        </button>
      `).join("")}
    </div>
  `;
  updateDocGuideToggle();
}

function setDocGuideOpen(open) {
  docGuideOpen = Boolean(open);
  renderDocGuide(activeSourceHeadings);
  bindDocGuideLinks();
}

function updateDocGuideToggle() {
  if (!docGuideToggle) return;
  const hasGuide = list(activeSourceHeadings).length > 0;
  docGuideToggle.disabled = !hasGuide;
  docGuideToggle.textContent = hasGuide
    ? t(docGuideOpen ? "source.guideHide" : "source.guideShow")
    : t("source.guideUnavailable");
  docGuideToggle.setAttribute("aria-expanded", String(hasGuide && docGuideOpen));
}

function bindDocGuideLinks() {
  if (!docGuide) return;
  docGuide.querySelectorAll("[data-heading-id]").forEach((button) => {
    button.addEventListener("click", () => {
      if (activeDocMode === "raw") {
        scrollRawSourceToLine(Number(button.dataset.headingLine || 1));
        return;
      }
      const target = document.getElementById(button.dataset.headingId);
      if (target) target.scrollIntoView({ block: "start", behavior: "smooth" });
    });
  });
}

function scrollRawSourceToLine(line) {
  const style = window.getComputedStyle(sourceContent);
  const lineHeight = Number.parseFloat(style.lineHeight) || 18;
  sourceContent.scrollTop = Math.max(0, (Math.max(line, 1) - 1) * lineHeight - 24);
}

function plainHeadingText(text) {
  return String(text || "")
    .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1")
    .replace(/[`*_#]/g, "")
    .trim();
}

async function renderMermaidBlocks() {
  const nodes = sourceRendered.querySelectorAll(".mermaid");
  if (nodes.length === 0 || typeof mermaid === "undefined") return;
  try {
    if (!mermaidReady) {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: "strict",
        theme: "default",
        flowchart: { htmlLabels: true }
      });
      mermaidReady = true;
    }
    await mermaid.run({ nodes });
  } catch (error) {
    nodes.forEach((node) => {
      node.classList.add("mermaid-error");
    });
  }
}

function splitFrontmatter(markdown) {
  const lines = markdown.split("\n");
  if (lines[0] !== "---") {
    return { frontmatter: [], body: markdown, bodyStartLine: 1 };
  }
  const end = lines.findIndex((line, index) => index > 0 && line === "---");
  if (end < 0) {
    return { frontmatter: [], body: markdown, bodyStartLine: 1 };
  }
  return {
    frontmatter: lines.slice(1, end),
    body: lines.slice(end + 1).join("\n"),
    bodyStartLine: end + 2
  };
}

function renderFrontmatter(lines) {
  const rows = [];
  let current = null;
  const pushCurrent = () => {
    if (!current) return;
    rows.push(current);
    current = null;
  };

  lines.forEach((line) => {
    const keyValue = /^([A-Za-z0-9_.-]+):\s*(.*)$/.exec(line);
    if (keyValue) {
      pushCurrent();
      current = { key: keyValue[1], values: keyValue[2] ? [keyValue[2]] : [] };
      return;
    }
    const listValue = /^\s*-\s+(.+)$/.exec(line);
    if (listValue && current) {
      current.values.push(listValue[1]);
      return;
    }
    if (line.trim() && current) {
      current.values.push(line.trim());
    }
  });
  pushCurrent();

  if (rows.length === 0) {
    return `<section class="frontmatter-block"><h2>${escapeHTML(t("frontmatter.title"))}</h2><pre><code>${escapeHTML(lines.join("\n"))}</code></pre></section>`;
  }
  return `
    <section class="frontmatter-block">
      <h2>${escapeHTML(t("frontmatter.title"))}</h2>
      <table>
        ${rows.map((row) => `<tr><th>${escapeHTML(row.key)}</th><td>${row.values.length > 0 ? row.values.map(renderInline).join("<br>") : escapeHTML(t("frontmatter.undeclared"))}</td></tr>`).join("")}
      </table>
    </section>
  `;
}

function renderInline(text) {
  const placeholders = [];
  let escaped = escapeHTML(text);
  escaped = escaped.replace(/`([^`]+)`/g, (_, code) => {
    const token = `@@CODE${placeholders.length}@@`;
    placeholders.push(`<code>${code}</code>`);
    return token;
  });
  escaped = escaped.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
  escaped = escaped.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_, label, href) => {
    const safeHref = escapeAttr(href);
    return `<a href="${safeHref}" data-doc-link="${safeHref}">${label}</a>`;
  });
  placeholders.forEach((value, index) => {
    escaped = escaped.replace(`@@CODE${index}@@`, value);
  });
  return escaped;
}

function isTableLine(line) {
  return line.includes("|") && line.startsWith("|") && line.endsWith("|");
}

function renderTable(lines) {
  if (lines.length < 2) {
    return lines.map((line) => `<p>${renderInline(line)}</p>`).join("");
  }
  const rows = lines.filter((line) => !/^\|\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)+\|?$/.test(line));
  if (rows.length === 0) return "";
  return `<table>${rows.map((line, rowIndex) => {
    const cells = line.split("|").slice(1, -1);
    const tag = rowIndex === 0 ? "th" : "td";
    return `<tr>${cells.map((cell) => `<${tag}>${renderInline(cell.trim())}</${tag}>`).join("")}</tr>`;
  }).join("")}</table>`;
}

function startSnapshotPolling() {
  window.setInterval(loadSnapshot, SNAPSHOT_POLL_INTERVAL_MS);
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function escapeAttr(value) {
  return escapeHTML(value).replaceAll("'", "&#39;");
}

function list(value) {
  return Array.isArray(value) ? value : [];
}

function bindRenderedDocLinks(currentPath) {
  sourceRendered.querySelectorAll("[data-doc-link]").forEach((link) => {
    link.addEventListener("click", (event) => {
      const targetPath = resolveDocLink(link.dataset.docLink, currentPath);
      if (!targetPath) return;
      event.preventDefault();
      navigateToSpecDocument(targetPath);
    });
  });
}

function resolveDocLink(rawHref, currentPath) {
  if (!rawHref) return "";
  const withoutHash = rawHref.split("#")[0];
  if (!withoutHash || /^[a-z]+:/i.test(withoutHash)) return "";
  if (!withoutHash.endsWith(".md")) return "";

  let resolved = withoutHash;
  if (!resolved.startsWith("docs/")) {
    const base = currentPath.split("/").slice(0, -1).join("/");
    resolved = normalizePath(`${base}/${resolved}`);
  }
  return normalizeSpecPath(resolved);
}

function normalizePath(path) {
  const stack = [];
  path.split("/").forEach((part) => {
    if (!part || part === ".") return;
    if (part === "..") {
      stack.pop();
      return;
    }
    stack.push(part);
  });
  return stack.join("/");
}

function normalizeSpecPath(path) {
  const normalized = normalizePath(path);
  const basename = normalized.split("/").pop();
  const direct = findKnownSpecPath(normalized);
  if (direct) return direct;

  const ruleMatch = /(?:^|\/)rule\/(candidate|stable)\/([^/]+\.md)$/.exec(normalized);
  if (ruleMatch) {
    const rulePath = `docs/specs/rules/${ruleMatch[1]}/${ruleMatch[2]}`;
    const knownRulePath = findKnownSpecPath(rulePath);
    if (knownRulePath) return knownRulePath;
  }

  return findKnownSpecPathByBasename(basename) || normalized;
}

function findKnownSpecPath(path) {
  if (isReadableOriginalPath(path)) {
    const exists = list(snapshot.sources).some((source) => source.path === path)
      || list(snapshot.nodes).some((node) => node.source && node.source.path === path);
    if (exists) return path;
  }
  return "";
}

function findKnownSpecPathByBasename(basename) {
  if (!basename) return "";
  const candidates = list(snapshot.sources)
    .map((source) => source.path)
    .concat(list(snapshot.nodes).map((node) => node.source && node.source.path).filter(Boolean))
    .filter((path) => path && path.endsWith(`/${basename}`));
  const unique = [...new Set(candidates)];
  return unique.length === 1 ? unique[0] : "";
}

function navigateToSpecDocument(path) {
  const graph = graphForCurrentView();
  const targetNode = graph.nodes.find((node) => node.source && node.source.path === path)
    || graph.nodes.find((node) => node.id === `file:${path}`);
  if (targetNode) {
    selectedNodeID = targetNode.id;
    renderNav();
    renderDetailForNode(targetNode.id);
    focusGraphNode(targetNode.id, 1.05);
  } else {
    const sourceObject = list(snapshot.objects).find((object) => list(object.truth_paths).some((ref) => ref.path === path));
    if (sourceObject) {
      selectedNodeID = objectNodeID(sourceObject);
      renderNav();
      renderDetailForNode(selectedNodeID);
      focusGraphNode(selectedNodeID, 1.05);
    }
  }
  openSource(path);
}

applyStaticText();
loadSnapshot();
startSnapshotPolling();
setDocMode("rendered");
