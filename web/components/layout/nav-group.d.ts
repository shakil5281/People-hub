import * as React from "react";
import { type LucideIcon } from "lucide-react";
export declare function NavGroup({ items, label, }: {
    items: {
        title: string;
        url: string;
        icon?: LucideIcon;
        items?: {
            title: string;
            url: string;
        }[];
    }[];
    label?: string;
}): React.JSX.Element;
