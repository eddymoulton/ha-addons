<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let vertical = true;

  const dispatch = createEventDispatcher<{
    resize: { deltaX: number; deltaY: number };
  }>();

  let isDragging = false;
  let startX = 0;
  let startY = 0;

  function handleMouseDown(event: MouseEvent) {
    isDragging = true;
    startX = event.clientX;
    startY = event.clientY;
    event.preventDefault();

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.cursor = vertical ? "col-resize" : "row-resize";
    document.body.style.userSelect = "none";
  }

  function handleMouseMove(event: MouseEvent) {
    if (!isDragging) return;

    const deltaX = event.clientX - startX;
    const deltaY = event.clientY - startY;

    dispatch("resize", { deltaX, deltaY });

    startX = event.clientX;
    startY = event.clientY;
  }

  function handleMouseUp() {
    isDragging = false;
    document.removeEventListener("mousemove", handleMouseMove);
    document.removeEventListener("mouseup", handleMouseUp);
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  }
</script>

<!-- svelte-ignore a11y_no_interactive_element_to_noninteractive_role -->
<button
  type="button"
  class="resize-handle"
  class:vertical
  class:horizontal={!vertical}
  class:dragging={isDragging}
  on:mousedown={handleMouseDown}
  role="separator"
  aria-label="Resize {vertical ? 'column' : 'row'}"
  aria-orientation={vertical ? "vertical" : "horizontal"}
></button>

<style>
  .resize-handle {
    position: relative;
    background: transparent;
    border: none;
    padding: 0;
    z-index: 10;
    flex-shrink: 0;
  }

  .resize-handle.vertical {
    width: 4px;
    cursor: col-resize;
    margin: 0 -2px;
  }

  .resize-handle.horizontal {
    height: 4px;
    cursor: row-resize;
    margin: -2px 0;
  }

  .resize-handle::before {
    content: "";
    position: absolute;
    transition: background-color 0.2s;
  }

  .resize-handle.vertical::before {
    width: 1px;
    height: 100%;
    left: 50%;
    transform: translateY(-50%);
    background: var(--ha-card-border-color, #2c2c2e);
  }

  .resize-handle.horizontal::before {
    height: 1px;
    width: 100%;
    top: 50%;
    transform: translateX(-50%);
    background: var(--ha-card-border-color, #2c2c2e);
  }

  .resize-handle:hover::before,
  .resize-handle.dragging::before {
    background: var(--primary-color, #03a9f4);
  }

  .resize-handle.vertical:hover::before,
  .resize-handle.vertical.dragging::before {
    width: 2px;
  }

  .resize-handle.horizontal:hover::before,
  .resize-handle.horizontal.dragging::before {
    height: 2px;
  }
</style>
