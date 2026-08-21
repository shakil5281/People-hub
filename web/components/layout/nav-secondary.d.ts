import * as React from "react";
import { SidebarGroup } from "@/components/ui/sidebar";
export declare function NavSecondary({ items, ...props }: {
    items: {
        title: string;
        url: string;
        icon: React.ReactNode;
    }[];
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>): React.JSX.Element;
