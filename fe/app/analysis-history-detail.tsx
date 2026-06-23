import React from 'react';
import {
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { getSelectedHistoryItem } from '@/lib/history-store';
import type { AnalysisHistoryResultJson } from '@/lib/api';

type DimensionKey = 'composition' | 'color' | 'exposure' | 'content';

const DIMENSION_META: { key: DimensionKey; label: string; icon: keyof typeof Ionicons.glyphMap }[] = [
  { key: 'composition', label: '构图', icon: 'grid-outline' },
  { key: 'color', label: '色彩', icon: 'color-palette-outline' },
  { key: 'exposure', label: '曝光', icon: 'sunny-outline' },
  { key: 'content', label: '内容识别', icon: 'scan-outline' },
];

function scoreColor(score: number): string {
  if (score >= 75) return '#5AE872';
  if (score >= 55) return '#F5C842';
  return '#FF6B6B';
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  const hh = String(d.getHours()).padStart(2, '0');
  const min = String(d.getMinutes()).padStart(2, '0');
  return `${d.getFullYear()}-${mm}-${dd}  ${hh}:${min}`;
}

export default function AnalysisHistoryDetailScreen() {
  const router = useRouter();
  const item = getSelectedHistoryItem();

  if (!item) {
    return (
      <SafeAreaView style={styles.root}>
        <View style={styles.header}>
          <TouchableOpacity onPress={() => router.back()} style={styles.backBtn} hitSlop={12}>
            <Ionicons name="arrow-back" size={20} color="#FFFFFF" />
          </TouchableOpacity>
          <Text style={styles.headerTitle}>详情</Text>
          <View style={{ width: 32 }} />
        </View>
        <View style={styles.center}>
          <Text style={styles.errorText}>记录已失效，请返回重试</Text>
        </View>
      </SafeAreaView>
    );
  }

  let parsed: AnalysisHistoryResultJson | null = null;
  try {
    parsed = JSON.parse(item.result_json) as AnalysisHistoryResultJson;
  } catch {
    // handled below
  }

  const DIMENSION_KEYS: DimensionKey[] = ['composition', 'color', 'exposure', 'content'];
  const avgScore =
    parsed
      ? Math.round(
          DIMENSION_KEYS.map((k) => parsed!.dimension_scores[k] ?? 0).reduce((a, b) => a + b, 0) / 4,
        )
      : 0;

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backBtn} hitSlop={12}>
          <Ionicons name="arrow-back" size={20} color="#FFFFFF" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>分析详情 #{item.id}</Text>
        <View style={{ width: 32 }} />
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Meta row */}
        <View style={styles.metaRow}>
          <Text style={styles.metaDate}>{formatDate(item.created_at)}</Text>
          <View style={[styles.scoreBig, { borderColor: scoreColor(avgScore) }]}>
            <Text style={[styles.scoreBigNum, { color: scoreColor(avgScore) }]}>{avgScore}</Text>
            <Text style={[styles.scoreBigLabel, { color: scoreColor(avgScore) }]}>综合</Text>
          </View>
        </View>

        {/* Dimension cards */}
        {parsed && (
          <>
            <Text style={styles.sectionTitle}>维度评分</Text>
            <View style={styles.dimGrid}>
              {DIMENSION_META.map((dim) => {
                const score = parsed!.dimension_scores[dim.key] ?? 0;
                const note = parsed!.dimension_notes[dim.key] ?? '';
                const color = scoreColor(score);
                return (
                  <View key={dim.key} style={styles.dimCard}>
                    <View style={styles.dimCardTop}>
                      <View style={styles.dimIconWrap}>
                        <Ionicons name={dim.icon} size={14} color="#555555" />
                      </View>
                      <Text style={styles.dimLabel}>{dim.label}</Text>
                      <Text style={[styles.dimScore, { color }]}>{score}</Text>
                    </View>
                    <View style={styles.dimBar}>
                      <View style={[styles.dimBarFill, { width: `${score}%` as any, backgroundColor: color }]} />
                    </View>
                    {!!note && <Text style={styles.dimNote}>{note}</Text>}
                  </View>
                );
              })}
            </View>

            {/* Overall analysis */}
            {!!parsed.overall_analysis && (
              <>
                <Text style={styles.sectionTitle}>综合分析</Text>
                <View style={styles.textCard}>
                  <Text style={styles.analysisText}>{parsed.overall_analysis}</Text>
                </View>
              </>
            )}

            {/* Focused deep analysis */}
            {!!parsed.focused_deep_analysis && (
              <>
                <Text style={styles.sectionTitle}>深度解析</Text>
                <View style={styles.textCard}>
                  <Text style={styles.analysisText}>{parsed.focused_deep_analysis}</Text>
                </View>
              </>
            )}

            {/* Improvement tips */}
            {parsed.improvement_tips?.length > 0 && (
              <>
                <Text style={styles.sectionTitle}>改进建议</Text>
                <View style={styles.tipsCard}>
                  {parsed.improvement_tips.map((tip, i) => (
                    <View key={i} style={styles.tipRow}>
                      <View style={styles.tipDot} />
                      <Text style={styles.tipText}>{tip}</Text>
                    </View>
                  ))}
                </View>
              </>
            )}
          </>
        )}

        {!parsed && (
          <View style={styles.center}>
            <Text style={styles.errorText}>数据解析失败</Text>
          </View>
        )}

        <View style={{ height: 40 }} />
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0A0A0A' },
  header: {
    height: 56,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  backBtn: { width: 32, alignItems: 'flex-start' },
  headerTitle: { fontFamily: 'PlayfairDisplay_700Bold', fontSize: 17, color: '#FFFFFF' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', paddingTop: 80 },
  errorText: { fontFamily: 'DMSans_400Regular', fontSize: 14, color: '#FF6B6B' },
  scroll: { flex: 1 },
  scrollContent: { padding: 20, gap: 0 },
  metaRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: 24,
  },
  metaDate: { fontFamily: 'DMMono_400Regular', fontSize: 12, color: '#444444' },
  scoreBig: {
    width: 56,
    height: 56,
    borderRadius: 28,
    borderWidth: 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scoreBigNum: { fontFamily: 'PlayfairDisplay_700Bold', fontSize: 20, lineHeight: 24 },
  scoreBigLabel: { fontFamily: 'DMMono_400Regular', fontSize: 9, lineHeight: 12 },
  sectionTitle: {
    fontFamily: 'DMSans_500Medium',
    fontSize: 13,
    color: '#555555',
    letterSpacing: 1,
    textTransform: 'uppercase',
    marginBottom: 10,
    marginTop: 20,
  },
  dimGrid: { gap: 8 },
  dimCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 14,
    gap: 8,
  },
  dimCardTop: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  dimIconWrap: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#1E1E1E',
    alignItems: 'center',
    justifyContent: 'center',
  },
  dimLabel: { fontFamily: 'DMSans_500Medium', fontSize: 14, color: '#FFFFFF', flex: 1 },
  dimScore: { fontFamily: 'DMMono_500Medium', fontSize: 16 },
  dimBar: {
    height: 3,
    backgroundColor: '#1E1E1E',
    borderRadius: 2,
    overflow: 'hidden',
  },
  dimBarFill: { height: 3, borderRadius: 2 },
  dimNote: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#777777',
    lineHeight: 20,
  },
  textCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 16,
  },
  analysisText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 14,
    color: '#CCCCCC',
    lineHeight: 22,
  },
  tipsCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 16,
    gap: 12,
  },
  tipRow: { flexDirection: 'row', gap: 10, alignItems: 'flex-start' },
  tipDot: {
    width: 4,
    height: 4,
    borderRadius: 2,
    backgroundColor: '#555555',
    marginTop: 8,
    flexShrink: 0,
  },
  tipText: { fontFamily: 'DMSans_400Regular', fontSize: 13, color: '#AAAAAA', lineHeight: 21, flex: 1 },
});
