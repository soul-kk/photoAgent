import React from 'react';
import {
  ActivityIndicator,
  FlatList,
  SafeAreaView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useAuth } from '@/context/auth';
import { fetchAnalysisHistory, type AnalysisHistoryItem, type AnalysisHistoryResultJson } from '@/lib/api';
import { setSelectedHistoryItem } from '@/lib/history-store';

const DIMENSION_KEYS = ['composition', 'color', 'exposure', 'content'] as const;

function avgScore(item: AnalysisHistoryItem): number {
  try {
    const parsed = JSON.parse(item.result_json) as AnalysisHistoryResultJson;
    const scores = DIMENSION_KEYS.map((k) => parsed.dimension_scores[k] ?? 0);
    return Math.round(scores.reduce((a, b) => a + b, 0) / scores.length);
  } catch {
    return 0;
  }
}

function briefText(item: AnalysisHistoryItem): string {
  try {
    const parsed = JSON.parse(item.result_json) as AnalysisHistoryResultJson;
    const text = parsed.overall_analysis ?? '';
    return text.length > 55 ? text.slice(0, 55) + '…' : text;
  } catch {
    return '暂无摘要';
  }
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  const hh = String(d.getHours()).padStart(2, '0');
  const min = String(d.getMinutes()).padStart(2, '0');
  return `${d.getFullYear()}-${mm}-${dd}  ${hh}:${min}`;
}

function scoreColor(score: number): string {
  if (score >= 75) return '#5AE872';
  if (score >= 55) return '#F5C842';
  return '#FF6B6B';
}

export default function AnalysisHistoryScreen() {
  const router = useRouter();
  const { token } = useAuth();

  const [items, setItems] = React.useState<AnalysisHistoryItem[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (!token) return;
    fetchAnalysisHistory(token)
      .then((data) => setItems(data.list))
      .catch((e) => setError(e instanceof Error ? e.message : '加载失败'))
      .finally(() => setLoading(false));
  }, [token]);

  function handlePress(item: AnalysisHistoryItem) {
    setSelectedHistoryItem(item);
    router.push('/analysis-history-detail' as any);
  }

  function renderItem({ item }: { item: AnalysisHistoryItem }) {
    const score = avgScore(item);
    const brief = briefText(item);
    return (
      <TouchableOpacity style={styles.card} onPress={() => handlePress(item)} activeOpacity={0.8}>
        <View style={styles.cardTop}>
          <View style={styles.cardMeta}>
            <Text style={styles.cardId}>#{item.id}</Text>
            <Text style={styles.cardType}>评分分析</Text>
          </View>
          <View style={[styles.scoreBadge, { borderColor: scoreColor(score) }]}>
            <Text style={[styles.scoreText, { color: scoreColor(score) }]}>{score}</Text>
          </View>
        </View>
        <Text style={styles.cardBrief} numberOfLines={2}>{brief}</Text>
        <Text style={styles.cardDate}>{formatDate(item.created_at)}</Text>
      </TouchableOpacity>
    );
  }

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => router.back()} style={styles.backBtn} hitSlop={12}>
          <Ionicons name="arrow-back" size={20} color="#FFFFFF" />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>AI 摄影分析记录</Text>
        <View style={{ width: 32 }} />
      </View>

      {loading && (
        <View style={styles.center}>
          <ActivityIndicator color="#FFFFFF" />
          <Text style={styles.loadingText}>加载中…</Text>
        </View>
      )}

      {!loading && error && (
        <View style={styles.center}>
          <Ionicons name="alert-circle-outline" size={36} color="#FF6B6B" />
          <Text style={styles.errorText}>{error}</Text>
        </View>
      )}

      {!loading && !error && items.length === 0 && (
        <View style={styles.center}>
          <Ionicons name="image-outline" size={48} color="#333333" />
          <Text style={styles.emptyText}>暂无分析记录</Text>
          <Text style={styles.emptySub}>去"分析"页面上传照片开始分析吧</Text>
        </View>
      )}

      {!loading && !error && items.length > 0 && (
        <FlatList
          data={items}
          keyExtractor={(item) => String(item.id)}
          renderItem={renderItem}
          contentContainerStyle={styles.list}
          showsVerticalScrollIndicator={false}
        />
      )}
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
  headerTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 17,
    color: '#FFFFFF',
  },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 12 },
  loadingText: { fontFamily: 'DMSans_400Regular', fontSize: 14, color: '#555555' },
  errorText: { fontFamily: 'DMSans_400Regular', fontSize: 14, color: '#FF6B6B', textAlign: 'center', paddingHorizontal: 32 },
  emptyText: { fontFamily: 'PlayfairDisplay_700Bold', fontSize: 18, color: '#333333' },
  emptySub: { fontFamily: 'DMSans_400Regular', fontSize: 13, color: '#444444', textAlign: 'center', paddingHorizontal: 32 },
  list: { padding: 16, gap: 12 },
  card: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 16,
    gap: 8,
  },
  cardTop: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  cardMeta: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  cardId: { fontFamily: 'DMMono_500Medium', fontSize: 13, color: '#FFFFFF' },
  cardType: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
    backgroundColor: '#1E1E1E',
    paddingHorizontal: 7,
    paddingVertical: 2,
    borderRadius: 3,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  scoreBadge: {
    width: 38,
    height: 38,
    borderRadius: 19,
    borderWidth: 1.5,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scoreText: { fontFamily: 'DMMono_500Medium', fontSize: 13 },
  cardBrief: { fontFamily: 'DMSans_400Regular', fontSize: 13, color: '#888888', lineHeight: 20 },
  cardDate: { fontFamily: 'DMMono_400Regular', fontSize: 11, color: '#444444' },
});
