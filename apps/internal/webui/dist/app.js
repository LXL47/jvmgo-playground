const presets = {
  hello: `public class Main {
    public static void main(String[] args) {
        String runtime = "JVMGO";
        System.out.println("Hello from " + runtime + "!");
        System.out.println("2 + 3 = " + (2 + 3));
    }
}`,
  fibonacci: `public class Main {
    public static void main(String[] args) {
        int a = 0;
        int b = 1;
        for (int i = 0; i < 12; i++) {
            System.out.println(a);
            int next = a + b;
            a = b;
            b = next;
        }
    }
}`,
  array: `public class Main {
    public static void main(String[] args) {
        int[] values = {42, 7, 19, 3, 28};
        for (int i = 0; i < values.length; i++) {
            for (int j = i + 1; j < values.length; j++) {
                if (values[j] < values[i]) {
                    int temp = values[i];
                    values[i] = values[j];
                    values[j] = temp;
                }
            }
        }
        for (int value : values) {
            System.out.println(value);
        }
    }
}`
};

const editor = document.querySelector("#sourceEditor");
const lineNumbers = document.querySelector("#lineNumbers");
const sourceStats = document.querySelector("#sourceStats");
const output = document.querySelector("#outputText");
const statusChip = document.querySelector("#statusChip");
const runButton = document.querySelector("#runButton");
const resetButton = document.querySelector("#resetButton");
const copyButton = document.querySelector("#copyButton");
const presetSelect = document.querySelector("#presetSelect");
const jobMeta = document.querySelector("#jobMeta");
const durationMeta = document.querySelector("#durationMeta");
const announcement = document.querySelector("#announcement");
const runtimeLight = document.querySelector("#runtimeLight");
const runtimeText = document.querySelector("#runtimeText");
const stages = [...document.querySelectorAll(".pipeline-stage")];
const connectors = [...document.querySelectorAll(".pipeline > i")];
let stageTimer = null;

function setSource(value) {
  editor.value = value;
  updateEditorMeta();
}

function updateEditorMeta() {
  const lines = editor.value.split("\n").length;
  lineNumbers.textContent = Array.from({ length: lines }, (_, index) => index + 1).join("\n");
  const bytes = new TextEncoder().encode(editor.value).length;
  sourceStats.textContent = `${lines} 行 · ${formatBytes(bytes)}`;
}

function setStatus(kind, label) {
  statusChip.className = `status-chip ${kind}`;
  statusChip.textContent = label;
  announcement.textContent = label;
}

function resetPipeline() {
  clearInterval(stageTimer);
  stages.forEach((stage, index) => {
    stage.classList.toggle("active", index === 0);
    stage.classList.remove("done");
  });
  connectors.forEach(connector => connector.classList.remove("done"));
}

function startPipeline() {
  resetPipeline();
  let current = 0;
  stageTimer = setInterval(() => {
    stages[current].classList.remove("active");
    stages[current].classList.add("done");
    if (connectors[current]) connectors[current].classList.add("done");
    current = Math.min(current + 1, stages.length - 1);
    stages[current].classList.add("active");
    if (current === stages.length - 1) clearInterval(stageTimer);
  }, 360);
}

function finishPipeline(success) {
  clearInterval(stageTimer);
  stages.forEach(stage => {
    stage.classList.remove("active");
    stage.classList.toggle("done", success);
  });
  connectors.forEach(connector => connector.classList.toggle("done", success));
}

async function loadRuntime() {
  try {
    const response = await fetch("/api/v1/runtime", { headers: { "Accept": "application/json" } });
    if (!response.ok) throw new Error("runtime unavailable");
    const limits = await response.json();
    document.querySelector("#instructionLimit").textContent = formatCount(limits.maxInstructions);
    document.querySelector("#heapLimit").textContent = formatBytes(limits.maxHeapBytes);
    document.querySelector("#arrayLimit").textContent = formatCount(limits.maxArrayLength);
    document.querySelector("#outputLimit").textContent = formatBytes(limits.maxOutputBytes);
    document.querySelector("#timeoutLimit").textContent = `${limits.timeoutMs / 1000} s`;
    runtimeLight.className = "state-light online";
    runtimeText.textContent = "运行时在线";
  } catch {
    runtimeLight.className = "state-light offline";
    runtimeText.textContent = "运行时离线";
  }
}

async function execute() {
  if (!editor.value.trim() || runButton.disabled) return;
  runButton.disabled = true;
  output.textContent = "正在编译并执行…";
  jobMeta.textContent = "正在创建任务";
  durationMeta.textContent = "-- ms";
  setStatus("running", "运行中");
  startPipeline();
  try {
    const response = await fetch("/api/v1/executions", {
      method: "POST",
      headers: { "Content-Type": "application/json", "Accept": "application/json" },
      body: JSON.stringify({ source: editor.value })
    });
    const data = await response.json();
    if (!response.ok && !data.status) throw new Error(data.error || "执行服务异常");
    output.textContent = data.output || "程序没有产生输出。";
    jobMeta.textContent = data.id ? `JOB ${data.id.slice(0, 8).toUpperCase()}` : "任务未创建";
    durationMeta.textContent = `${data.durationMs ?? 0} ms`;
    const success = data.status === "success";
    setStatus(success ? "success" : "error", statusLabel(data.status));
    finishPipeline(success);
  } catch (error) {
    output.textContent = error.message || "无法连接执行服务。";
    jobMeta.textContent = "请求失败";
    setStatus("error", "服务异常");
    finishPipeline(false);
  } finally {
    runButton.disabled = false;
  }
}

function statusLabel(status) {
  const labels = {
    success: "执行成功",
    compile_error: "编译错误",
    runtime_error: "运行错误",
		sandbox_limit: "预算终止",
    compile_timeout: "编译超时",
    timeout: "执行超时",
    output_limit: "输出超限",
    busy: "服务繁忙",
    invalid_request: "输入无效",
    internal_error: "内部错误"
  };
  return labels[status] || "执行失败";
}

function formatBytes(bytes) {
  if (bytes >= 1024 * 1024) return `${trimNumber(bytes / 1024 / 1024)} MiB`;
  if (bytes >= 1024) return `${trimNumber(bytes / 1024)} KiB`;
  return `${bytes} B`;
}

function formatCount(value) {
  if (value >= 1000000) return `${trimNumber(value / 1000000)} M`;
  if (value >= 1000) return `${trimNumber(value / 1000)} K`;
  return String(value);
}

function trimNumber(value) {
  return Number.isInteger(value) ? String(value) : value.toFixed(1);
}

editor.addEventListener("input", updateEditorMeta);
editor.addEventListener("scroll", () => { lineNumbers.scrollTop = editor.scrollTop; });
editor.addEventListener("keydown", event => {
  if (event.key === "Tab") {
    event.preventDefault();
    const start = editor.selectionStart;
    editor.setRangeText("    ", start, editor.selectionEnd, "end");
    updateEditorMeta();
  }
  if ((event.ctrlKey || event.metaKey) && event.key === "Enter") execute();
});
runButton.addEventListener("click", execute);
resetButton.addEventListener("click", () => {
  setSource(presets[presetSelect.value]);
  output.textContent = "等待执行。";
  jobMeta.textContent = "任务尚未创建";
  durationMeta.textContent = "-- ms";
  setStatus("ready", "就绪");
  resetPipeline();
});
presetSelect.addEventListener("change", () => setSource(presets[presetSelect.value]));
copyButton.addEventListener("click", async () => {
  await navigator.clipboard.writeText(output.textContent);
  copyButton.textContent = "已复制";
  setTimeout(() => { copyButton.textContent = "复制输出"; }, 1200);
});

setSource(presets.hello);
resetPipeline();
loadRuntime();
