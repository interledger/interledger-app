import { unified } from 'unified'
import rehypeParse from 'rehype-parse'
import rehypeStringify from 'rehype-stringify'
import { read } from 'to-vfile'
import fs, { writeFileSync, mkdir } from 'fs'
import path from 'path'
import { visit } from 'unist-util-visit'

main()

async function main() {
  let templates = []
  for await (const p of walk('./templates')) {
    const extIndex = p.lastIndexOf('.')
    const ext = p.slice(extIndex)
    if (ext !== '.html') {
      continue
    }
    templates.push(p.slice(9))
    const index = p.lastIndexOf('/')
    mkdir(`./production${p.slice(9, index)}`, { recursive: true }, (err) => {
      if (err) throw err
    })
  }

  for (const temp of templates) {
    console.log(temp)
    const out = await unified()
      .use(rehypeParse, { fragment: true })
      .use(wrapTemplate)
      .use(rehypeStringify)
      .process(await read(`./templates${temp}`))
    writeFileSync(`./production${temp}`, out.value)
  }
}

async function* walk(dir) {
  for await (const d of await fs.promises.opendir(dir)) {
    const entry = path.join(dir, d.name)
    if (d.isDirectory()) yield* walk(entry)
    else if (d.isFile()) yield entry
  }
}

function wrapTemplate() {
  return async (tree, file) => {
    let layoutName = './layouts/main.html'
    visit(tree, { tagName: 'template' }, (node) => {
      layoutName = `./layouts/${node.properties.src}.html`
    })

    const layoutProcessor = await unified().use(rehypeParse, {
      fragment: false
    })

    const layout = layoutProcessor.parse(await read(layoutName))
    visit(layout, { tagName: 'article' }, (node) => {
      node.children = tree.children[0].content.children
    })

    const children = layout.children
    return {
      ...tree,
      children: Array.isArray(children) ? children : [children]
    }
  }
}
