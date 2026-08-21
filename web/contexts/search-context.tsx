"use client"

import * as React from "react"

interface SearchContextType {
  open: boolean
  setOpen: (open: boolean) => void
}

const SearchContext = React.createContext<SearchContextType>({
  open: false,
  setOpen: () => {},
})

export function SearchProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = React.useState(false)

  React.useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((prev) => !prev)
      }
    }
    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [])

  return (
    <SearchContext.Provider value={{ open, setOpen }}>
      {children}
    </SearchContext.Provider>
  )
}

export function useSearchDialog() {
  return React.useContext(SearchContext)
}
