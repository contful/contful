// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

/**
 * Contful Console — ESLint Flat Config (v10+)
 * 技术栈: Vue 3 + TypeScript + TDesign
 *
 * 规则策略：
 *   - JS 基础推荐规则（js.configs.recommended）
 *   - TypeScript 推荐规则（typescript-eslint recommended）
 *   - Vue 3 Essential 规则（防止错误的规则，不含风格规则）
 */

import tseslint from 'typescript-eslint'
import pluginVue from 'eslint-plugin-vue'
import js from '@eslint/js'
import globals from 'globals'

export default [
  // 基础 JS 推荐规则
  js.configs.recommended,

  // TypeScript 推荐规则
  ...tseslint.configs.recommended,

  // Vue 3 Essential 规则（仅防错，不含风格规则）
  ...pluginVue.configs['flat/essential'],

  // 全局配置
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
    },
    rules: {
      // TypeScript —— 宽松
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-empty-object-type': 'off',
      '@typescript-eslint/no-unused-vars': 'warn',

      // Vue
      'vue/multi-word-component-names': 'off',

      // 通用
      'no-unused-vars': 'warn',
      'no-undef': 'warn',
    },
  },

  // Vue SFC — TypeScript parser
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },

  // TS 独立文件
  {
    files: ['**/*.{ts,mts,cts,tsx}'],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project: './tsconfig.json',
      },
    },
  },

  // 忽略目录
  {
    ignores: [
      'dist/',
      'node_modules/',
      'public/',
      '*.min.js',
    ],
  },
]
