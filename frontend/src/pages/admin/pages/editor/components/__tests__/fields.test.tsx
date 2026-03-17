import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import TextField from "../fields/TextField";
import TextareaField from "../fields/TextareaField";
import NumberField from "../fields/NumberField";
import BooleanField from "../fields/BooleanField";
import ColorField from "../fields/ColorField";
import SelectField from "../fields/SelectField";
import BilingualField from "../fields/BilingualField";
import BilingualTextareaField from "../fields/BilingualTextareaField";
import MediaField from "../fields/MediaField";
import StringArrayField from "../fields/StringArrayField";

// Mock ImagePickerModal to avoid complex dependencies in tests
vi.mock("@/components/admin/ImagePickerModal", () => ({
  default: ({ open, onClose, onSelect }: any) =>
    open ? (
      <div data-testid="image-picker-modal">
        <button onClick={onClose}>close-modal</button>
        <button onClick={() => onSelect({ url: "/picked.jpg" })}>pick-image</button>
      </div>
    ) : null,
}));

describe("TextField", () => {
  it("renders value and calls onChange", () => {
    const onChange = vi.fn();
    render(
      <TextField
        schema={{ key: "t", type: "text", label: "标题" }}
        value="hello"
        onChange={onChange}
      />
    );
    expect(screen.getByDisplayValue("hello")).toBeInTheDocument();
    fireEvent.change(screen.getByDisplayValue("hello"), {
      target: { value: "world" },
    });
    expect(onChange).toHaveBeenCalledWith("world");
  });

  it("renders label", () => {
    render(
      <TextField
        schema={{ key: "t", type: "text", label: "标题" }}
        value=""
        onChange={vi.fn()}
      />
    );
    expect(screen.getByText("标题")).toBeInTheDocument();
  });

  it("handles null value", () => {
    render(
      <TextField
        schema={{ key: "t", type: "text", label: "标题" }}
        value={null}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByRole("textbox")).toHaveValue("");
  });
});

describe("TextareaField", () => {
  it("renders value and calls onChange", () => {
    const onChange = vi.fn();
    render(
      <TextareaField
        schema={{ key: "d", type: "textarea", label: "描述" }}
        value="some text"
        onChange={onChange}
      />
    );
    expect(screen.getByDisplayValue("some text")).toBeInTheDocument();
    fireEvent.change(screen.getByDisplayValue("some text"), {
      target: { value: "new text" },
    });
    expect(onChange).toHaveBeenCalledWith("new text");
  });

  it("renders label", () => {
    render(
      <TextareaField
        schema={{ key: "d", type: "textarea", label: "描述" }}
        value=""
        onChange={vi.fn()}
      />
    );
    expect(screen.getByText("描述")).toBeInTheDocument();
  });
});

describe("NumberField", () => {
  it("renders numeric value", () => {
    render(
      <NumberField
        schema={{ key: "n", type: "number", label: "数量" }}
        value={42}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("42")).toBeInTheDocument();
  });

  it("calls onChange with number", () => {
    const onChange = vi.fn();
    render(
      <NumberField
        schema={{ key: "n", type: "number", label: "数量" }}
        value={10}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("10"), {
      target: { value: "20" },
    });
    expect(onChange).toHaveBeenCalledWith(20);
  });

  it("calls onChange with undefined when empty", () => {
    const onChange = vi.fn();
    render(
      <NumberField
        schema={{ key: "n", type: "number", label: "数量" }}
        value={10}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("10"), {
      target: { value: "" },
    });
    expect(onChange).toHaveBeenCalledWith(undefined);
  });
});

describe("BooleanField", () => {
  it("renders checked state", () => {
    render(
      <BooleanField
        schema={{ key: "b", type: "boolean", label: "启用" }}
        value={true}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByRole("checkbox")).toBeChecked();
  });

  it("renders label inline", () => {
    render(
      <BooleanField
        schema={{ key: "b", type: "boolean", label: "启用" }}
        value={false}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByText("启用")).toBeInTheDocument();
    expect(screen.getByRole("checkbox")).not.toBeChecked();
  });

  it("calls onChange with boolean", () => {
    const onChange = vi.fn();
    render(
      <BooleanField
        schema={{ key: "b", type: "boolean", label: "启用" }}
        value={false}
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getByRole("checkbox"));
    expect(onChange).toHaveBeenCalledWith(true);
  });
});

describe("ColorField", () => {
  it("renders color value and swatch", () => {
    render(
      <ColorField
        schema={{ key: "c", type: "color", label: "颜色" }}
        value="#ff0000"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("#ff0000")).toBeInTheDocument();
    expect(screen.getByText("颜色")).toBeInTheDocument();
  });

  it("calls onChange with string", () => {
    const onChange = vi.fn();
    render(
      <ColorField
        schema={{ key: "c", type: "color", label: "颜色" }}
        value="#000"
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("#000"), {
      target: { value: "#fff" },
    });
    expect(onChange).toHaveBeenCalledWith("#fff");
  });
});

describe("SelectField", () => {
  it("renders options", () => {
    render(
      <SelectField
        schema={{
          key: "s",
          type: "select",
          label: "布局",
          options: [
            { label: "左", value: "left" },
            { label: "右", value: "right" },
          ],
        }}
        value="left"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByRole("combobox")).toHaveValue("left");
    expect(screen.getByText("请选择")).toBeInTheDocument();
  });

  it("coerces numeric option values", () => {
    const onChange = vi.fn();
    render(
      <SelectField
        schema={{
          key: "col",
          type: "select",
          label: "列数",
          options: [
            { label: "2列", value: 2 },
            { label: "3列", value: 3 },
          ],
        }}
        value={2}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "3" },
    });
    expect(onChange).toHaveBeenCalledWith(3); // number not string!
  });

  it("passes string values for non-numeric options", () => {
    const onChange = vi.fn();
    render(
      <SelectField
        schema={{
          key: "s",
          type: "select",
          label: "布局",
          options: [
            { label: "左", value: "left" },
            { label: "右", value: "right" },
          ],
        }}
        value="left"
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByRole("combobox"), {
      target: { value: "right" },
    });
    expect(onChange).toHaveBeenCalledWith("right");
  });
});

describe("BilingualField", () => {
  it("renders zh tab by default", () => {
    render(
      <BilingualField
        schema={{ key: "t", type: "bilingual", label: "标题" }}
        value={{ zh: "你好", en: "Hello" }}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("你好")).toBeInTheDocument();
  });

  it("switches to en tab", () => {
    render(
      <BilingualField
        schema={{ key: "t", type: "bilingual", label: "标题" }}
        value={{ zh: "你好", en: "Hello" }}
        onChange={vi.fn()}
      />
    );
    fireEvent.click(screen.getByText("en"));
    expect(screen.getByDisplayValue("Hello")).toBeInTheDocument();
  });

  it("handles legacy string value", () => {
    render(
      <BilingualField
        schema={{ key: "t", type: "bilingual", label: "标题" }}
        value="旧数据"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("旧数据")).toBeInTheDocument();
  });

  it("onChange emits {zh, en} object", () => {
    const onChange = vi.fn();
    render(
      <BilingualField
        schema={{ key: "t", type: "bilingual", label: "标题" }}
        value={{ zh: "你好", en: "Hello" }}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("你好"), {
      target: { value: "世界" },
    });
    expect(onChange).toHaveBeenCalledWith({ zh: "世界", en: "Hello" });
  });

  it("handles null value", () => {
    render(
      <BilingualField
        schema={{ key: "t", type: "bilingual", label: "标题" }}
        value={null}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByRole("textbox")).toHaveValue("");
  });
});

describe("BilingualTextareaField", () => {
  it("renders zh tab by default", () => {
    render(
      <BilingualTextareaField
        schema={{ key: "d", type: "bilingual-textarea", label: "描述" }}
        value={{ zh: "中文描述", en: "English desc" }}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("中文描述")).toBeInTheDocument();
  });

  it("switches to en tab", () => {
    render(
      <BilingualTextareaField
        schema={{ key: "d", type: "bilingual-textarea", label: "描述" }}
        value={{ zh: "中文", en: "English" }}
        onChange={vi.fn()}
      />
    );
    fireEvent.click(screen.getByText("en"));
    expect(screen.getByDisplayValue("English")).toBeInTheDocument();
  });

  it("onChange emits {zh, en} object", () => {
    const onChange = vi.fn();
    render(
      <BilingualTextareaField
        schema={{ key: "d", type: "bilingual-textarea", label: "描述" }}
        value={{ zh: "原文", en: "orig" }}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("原文"), {
      target: { value: "新文" },
    });
    expect(onChange).toHaveBeenCalledWith({ zh: "新文", en: "orig" });
  });

  it("handles legacy string value", () => {
    render(
      <BilingualTextareaField
        schema={{ key: "d", type: "bilingual-textarea", label: "描述" }}
        value="旧文本"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("旧文本")).toBeInTheDocument();
  });
});

describe("MediaField", () => {
  it("renders URL input", () => {
    render(
      <MediaField
        schema={{ key: "img", type: "media", label: "图片" }}
        value="/test.jpg"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("/test.jpg")).toBeInTheDocument();
  });

  it("shows thumbnail when value present", () => {
    render(
      <MediaField
        schema={{ key: "img", type: "media", label: "图片" }}
        value="/test.jpg"
        onChange={vi.fn()}
      />
    );
    expect(screen.getByAltText("preview")).toHaveAttribute("src", "/test.jpg");
  });

  it("shows 选择 button", () => {
    render(
      <MediaField
        schema={{ key: "img", type: "media", label: "图片" }}
        value=""
        onChange={vi.fn()}
      />
    );
    expect(screen.getByText("选择")).toBeInTheDocument();
  });

  it("does not show thumbnail when value is empty", () => {
    render(
      <MediaField
        schema={{ key: "img", type: "media", label: "图片" }}
        value=""
        onChange={vi.fn()}
      />
    );
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("opens picker and selects image", () => {
    const onChange = vi.fn();
    render(
      <MediaField
        schema={{ key: "img", type: "media", label: "图片" }}
        value=""
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getByText("选择"));
    expect(screen.getByTestId("image-picker-modal")).toBeInTheDocument();
    fireEvent.click(screen.getByText("pick-image"));
    expect(onChange).toHaveBeenCalledWith("/picked.jpg");
  });
});

describe("StringArrayField", () => {
  it("renders each string as input", () => {
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A", "B"]}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByDisplayValue("A")).toBeInTheDocument();
    expect(screen.getByDisplayValue("B")).toBeInTheDocument();
  });

  it("adds empty item", () => {
    const onChange = vi.fn();
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A"]}
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getByText("+ 添加"));
    expect(onChange).toHaveBeenCalledWith(["A", ""]);
  });

  it("deletes an item", () => {
    const onChange = vi.fn();
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A", "B"]}
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getAllByTitle("删除")[0]);
    expect(onChange).toHaveBeenCalledWith(["B"]);
  });

  it("moves item up", () => {
    const onChange = vi.fn();
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A", "B", "C"]}
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getAllByTitle("上移")[1]);
    expect(onChange).toHaveBeenCalledWith(["B", "A", "C"]);
  });

  it("moves item down", () => {
    const onChange = vi.fn();
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A", "B", "C"]}
        onChange={onChange}
      />
    );
    fireEvent.click(screen.getAllByTitle("下移")[0]);
    expect(onChange).toHaveBeenCalledWith(["B", "A", "C"]);
  });

  it("updates an item value", () => {
    const onChange = vi.fn();
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={["A", "B"]}
        onChange={onChange}
      />
    );
    fireEvent.change(screen.getByDisplayValue("A"), {
      target: { value: "X" },
    });
    expect(onChange).toHaveBeenCalledWith(["X", "B"]);
  });

  it("handles non-array value gracefully", () => {
    render(
      <StringArrayField
        schema={{ key: "items", type: "string-array", label: "项" }}
        value={null}
        onChange={vi.fn()}
      />
    );
    expect(screen.getByText("+ 添加")).toBeInTheDocument();
  });
});
