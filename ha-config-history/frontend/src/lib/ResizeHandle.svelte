<script lang="ts">
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher<{
    resize: { deltaX: number };
  }>();

  let isDragging = $state(false);
  let startX = 0;

  function handleMouseDown(event: MouseEvent) {
    isDragging = true;
    startX = event.clientX;
    event.preventDefault();

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
  }

  function handleMouseMove(event: MouseEvent) {
    if (!isDragging) return;

    const deltaX = event.clientX - startX;

    dispatch("resize", { deltaX });

    startX = event.clientX;
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
  class:dragging={isDragging}
  onmousedown={handleMouseDown}
  role="separator"
  aria-label="Resize column"
  aria-orientation="vertical"
></button>

<style>
  .resize-handle {
    position: relative;
    background: transparent;
    border: none;
    padding: 0;
    z-index: 10;
    flex-shrink: 0;
    width: 4px;
    cursor: col-resize;
    margin: 0 -2px;
  }

  .resize-handle::before {
    content: "";
    position: absolute;
    width: 1px;
    height: 100%;
    left: 50%;
    transform: translateY(-50%);
    background: var(--ha-card-border-color);
    transition: background-color 0.2s;
  }

  .resize-handle:hover::before,
  .resize-handle.dragging::before {
    background: var(--primary-color);
    width: 2px;
  }
</style>
