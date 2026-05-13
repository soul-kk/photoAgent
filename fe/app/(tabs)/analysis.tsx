import React from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { TabBar } from '@/components/tab-bar';

const DIMS = ['构图', '色彩', '曝光', '内容识别'];

const SCORES = [
  { label: '构图', value: 87, pct: 0.87 },
  { label: '色彩', value: 72, pct: 0.72 },
  { label: '曝光', value: 91, pct: 0.91 },
];

export default function AnalysisScreen() {
  const [activeDim, setActiveDim] = React.useState(0);

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>AI 摄影分析</Text>
        <Text style={styles.headerSub}>上传照片开始分析</Text>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Dropzone */}
        <View style={styles.dropzone}>
          <Ionicons name="cloud-upload-outline" size={40} color="#555555" />
          <Text style={styles.dropzoneTitle}>上传照片</Text>
          <Text style={styles.dropzoneSub}>JPG · PNG · HEIF · 最大 20MB</Text>
        </View>

        {/* Analyze Button */}
        <TouchableOpacity style={styles.analyzeBtn} activeOpacity={0.85}>
          <Text style={styles.analyzeBtnText}>开始 AI 分析</Text>
        </TouchableOpacity>

        {/* Dimension Tags */}
        <Text style={styles.sectionLabel}>分析维度</Text>
        <View style={styles.dimsRow}>
          {DIMS.map((d, i) => (
            <TouchableOpacity
              key={d}
              style={[styles.dimTag, activeDim === i && styles.dimTagActive]}
              onPress={() => setActiveDim(i)}
              activeOpacity={0.8}
            >
              <Text style={[styles.dimTagText, activeDim === i && styles.dimTagTextActive]}>
                {d}
              </Text>
            </TouchableOpacity>
          ))}
        </View>

        <View style={styles.divider} />

        {/* Results */}
        <Text style={styles.sectionLabel}>分析结果</Text>
        <View style={styles.scoresRow}>
          {SCORES.map((s) => (
            <View key={s.label} style={styles.scoreCard}>
              <Text style={styles.scoreLabel}>{s.label}</Text>
              <Text style={styles.scoreValue}>{s.value}</Text>
              <View style={styles.scoreBarBg}>
                <View style={[styles.scoreBarFill, { width: `${s.pct * 100}%` as any }]} />
              </View>
            </View>
          ))}
        </View>

        {/* Report Card */}
        <View style={styles.reportCard}>
          <Text style={styles.reportLabel}>AI 分析报告</Text>
          <Text style={styles.reportText}>
            该照片采用经典三分法构图，主体位于画面右侧黄金分割点，视觉引导自然流畅。前景虚化处理有效突出主体，景深控制恰当。
          </Text>
          <View style={styles.reportDivider} />
          <Text style={styles.reportSuggest}>
            建议：适当提升阴影细节，右下角有轻微欠曝，建议后期补偿 +0.5EV。色彩饱和度偏低，可提升整体对比度以增强视觉冲击力。
          </Text>
        </View>
      </ScrollView>

      <TabBar />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  root: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  header: {
    height: 56,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  headerTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 18,
    color: '#FFFFFF',
  },
  headerSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
  },
  scroll: {
    flex: 1,
  },
  scrollContent: {
    padding: 24,
    paddingBottom: 120,
    gap: 24,
  },
  dropzone: {
    height: 220,
    backgroundColor: '#111111',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#333333',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
  },
  dropzoneTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 18,
    color: '#FFFFFF',
  },
  dropzoneSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
  },
  analyzeBtn: {
    height: 52,
    backgroundColor: '#FFFFFF',
    borderRadius: 4,
    alignItems: 'center',
    justifyContent: 'center',
  },
  analyzeBtnText: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 15,
    color: '#0A0A0A',
  },
  sectionLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  dimsRow: {
    flexDirection: 'row',
    gap: 8,
    flexWrap: 'wrap',
  },
  dimTag: {
    backgroundColor: '#1E1E1E',
    borderWidth: 1,
    borderColor: '#333333',
    borderRadius: 2,
    paddingHorizontal: 14,
    paddingVertical: 7,
  },
  dimTagActive: {
    backgroundColor: '#FFFFFF',
    borderColor: '#FFFFFF',
  },
  dimTagText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 12,
    color: '#999999',
  },
  dimTagTextActive: {
    fontFamily: 'DMSans_700Bold',
    color: '#0A0A0A',
  },
  divider: {
    height: 1,
    backgroundColor: '#2A2A2A',
  },
  scoresRow: {
    flexDirection: 'row',
    gap: 8,
  },
  scoreCard: {
    flex: 1,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 14,
    paddingVertical: 16,
    gap: 8,
  },
  scoreLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 1,
  },
  scoreValue: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 32,
    color: '#FFFFFF',
  },
  scoreBarBg: {
    height: 2,
    backgroundColor: '#2A2A2A',
  },
  scoreBarFill: {
    height: 2,
    backgroundColor: '#FFFFFF',
  },
  reportCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    paddingVertical: 20,
    gap: 12,
  },
  reportLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  reportText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#FFFFFF',
    lineHeight: 22,
  },
  reportDivider: {
    height: 1,
    backgroundColor: '#2A2A2A',
  },
  reportSuggest: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#999999',
    lineHeight: 22,
  },
});
