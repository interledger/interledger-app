import type { RefObject } from 'react'
import { create } from 'zustand'

export type LoaderDocsNav = {
  id: string
  position?: number
  sections: Omit<DocSections, 'ref'>[]
  slug?: string
  title?: string
}

export type DocsNav = {
  id: string
  position?: number
  sections: DocSections[]
  slug?: string
  title?: string
}

type DocSections = {
  id: string
  slug?: string
  title?: string
  ref?: RefObject<HTMLHeadingElement>
}

interface DocsState {
  sections: DocsNav[]
  visibleSections: string[]
}

interface DocsActions {
  setSections: (sections: LoaderDocsNav[]) => void
  setVisibleSections: (visibleSections: string[]) => void
  registerHeading: (
    id: string,
    docId: string,
    ref: RefObject<HTMLHeadingElement>
  ) => void
  reset: () => void
}

const docsInitialState = {
  sections: [],
  visibleSections: []
}

export const useDocsStore = create<DocsState & DocsActions>()((set) => ({
  ...docsInitialState,
  setSections: (sections) =>
    set((state) => {
      return {
        sections: sections
          .map((doc) => {
            if (doc.title && doc.slug) {
              return {
                ...doc,
                sections: doc.sections.filter((section) => {
                  return section.title && section.slug
                })
              }
            }
            return doc
          })
          .filter((doc) => doc.title && doc.slug)
      }
    }),
  setVisibleSections: (visibleSections) =>
    set((state) =>
      state.visibleSections.join() === visibleSections.join()
        ? {}
        : { visibleSections }
    ),
  registerHeading: (id, docId, ref) =>
    set((state) => {
      return {
        sections: state.sections.map((doc) => {
          if (doc.id === docId) {
            return {
              ...doc,
              sections: doc.sections.map((section) => {
                if (section.id === id) {
                  return {
                    ...section,
                    ref
                  }
                }
                return section
              })
            }
          }
          return doc
        })
      }
    }),
  reset: () => set((state) => ({ ...docsInitialState }))
}))
