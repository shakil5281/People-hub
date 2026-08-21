import * as React from "react";
import { ContextMenu as ContextMenuPrimitive } from "radix-ui";
declare function ContextMenu({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Root>): React.JSX.Element;
declare function ContextMenuTrigger({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Trigger>): React.JSX.Element;
declare function ContextMenuGroup({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Group>): React.JSX.Element;
declare function ContextMenuPortal({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Portal>): React.JSX.Element;
declare function ContextMenuSub({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Sub>): React.JSX.Element;
declare function ContextMenuRadioGroup({ ...props }: React.ComponentProps<typeof ContextMenuPrimitive.RadioGroup>): React.JSX.Element;
declare function ContextMenuContent({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Content> & {
    side?: "top" | "right" | "bottom" | "left";
}): React.JSX.Element;
declare function ContextMenuItem({ className, inset, variant, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Item> & {
    inset?: boolean;
    variant?: "default" | "destructive";
}): React.JSX.Element;
declare function ContextMenuSubTrigger({ className, inset, children, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.SubTrigger> & {
    inset?: boolean;
}): React.JSX.Element;
declare function ContextMenuSubContent({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.SubContent>): React.JSX.Element;
declare function ContextMenuCheckboxItem({ className, children, checked, inset, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.CheckboxItem> & {
    inset?: boolean;
}): React.JSX.Element;
declare function ContextMenuRadioItem({ className, children, inset, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.RadioItem> & {
    inset?: boolean;
}): React.JSX.Element;
declare function ContextMenuLabel({ className, inset, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Label> & {
    inset?: boolean;
}): React.JSX.Element;
declare function ContextMenuSeparator({ className, ...props }: React.ComponentProps<typeof ContextMenuPrimitive.Separator>): React.JSX.Element;
declare function ContextMenuShortcut({ className, ...props }: React.ComponentProps<"span">): React.JSX.Element;
export { ContextMenu, ContextMenuTrigger, ContextMenuContent, ContextMenuItem, ContextMenuCheckboxItem, ContextMenuRadioItem, ContextMenuLabel, ContextMenuSeparator, ContextMenuShortcut, ContextMenuGroup, ContextMenuPortal, ContextMenuSub, ContextMenuSubContent, ContextMenuSubTrigger, ContextMenuRadioGroup, };
