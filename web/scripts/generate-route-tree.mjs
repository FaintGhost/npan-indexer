import process from 'node:process'
import { Generator, getConfig } from '@tanstack/router-generator'

const root = process.cwd()

const config = getConfig(
  {
    target: 'react',
    autoCodeSplitting: true,
    routesDirectory: './src/routes',
    generatedRouteTree: './src/routeTree.gen.ts',
    routeFileIgnorePattern: '\\.(test|spec)\\.(ts|tsx)$',
  },
  root,
)

const generator = new Generator({ config, root })
await generator.run()
