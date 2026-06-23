import React from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  SafeAreaView,
} from 'react-native';
import { useRouter } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { TabBar } from '@/components/tab-bar';

export default function HomeScreen() {
  const router = useRouter();

  return (
    <SafeAreaView style={styles.root}>
      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Hero */}
        <View style={styles.hero}>
          <Text style={styles.eyebrow}>AI PHOTOGRAPHY</Text>
          <Text style={styles.headline}>LENS AI</Text>
          <Text style={styles.tagline}>让每一张照片都值得被记住</Text>
        </View>

        {/* Cards */}
        <View style={styles.cards}>
          <TouchableOpacity
            style={[styles.card, styles.cardBorderBottom]}
            onPress={() => router.push('./analysis')}
            activeOpacity={0.8}
          >
            <Ionicons name="scan-outline" size={24} color="#FFFFFF" />
            <View style={styles.cardRow}>
              <Text style={styles.cardTitle}>AI 摄影分析</Text>
              <Text style={styles.cardArrow}>→</Text>
            </View>
            <Text style={styles.cardDesc}>
              构图 · 色彩 · 曝光全方位解析，获得专业改进建议
            </Text>
            <View style={styles.tag}>
              <Text style={styles.tagText}>核心功能</Text>
            </View>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.card, styles.cardBorderBottom]}
            onPress={() => router.push('./advice')}
            activeOpacity={0.8}
          >
            <Ionicons name="camera-outline" size={24} color="#FFFFFF" />
            <View style={styles.cardRow}>
              <Text style={styles.cardTitle}>AI 拍摄建议</Text>
              <Text style={styles.cardArrow}>→</Text>
            </View>
            <Text style={styles.cardDesc}>
              拍前规划 · 机位 · 焦段建议，让每次出片都有方向
            </Text>
            <View style={styles.tag}>
              <Text style={styles.tagText}>核心功能</Text>
            </View>
          </TouchableOpacity>

          {/* <TouchableOpacity
            style={[styles.card, styles.cardBorderBottom]}
            onPress={() => router.push('/compare')}
            activeOpacity={0.8}
          >
            <Ionicons name="images-outline" size={24} color="#FFFFFF" />
            <View style={styles.cardRow}>
              <Text style={styles.cardTitle}>图片质量判断</Text>
              <Text style={styles.cardArrow}>→</Text>
            </View>
            <Text style={styles.cardDesc}>
              多图对比 · 智能排序 · 自动推荐最佳照片
            </Text>
            <View style={[styles.tag, styles.tagExtend]}>
              <Text style={styles.tagText}>拓展功能</Text>
            </View>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.card}
            onPress={() => router.push('/tone')}
            activeOpacity={0.8}
          >
            <Ionicons name="color-palette-outline" size={24} color="#FFFFFF" />
            <View style={styles.cardRow}>
              <Text style={styles.cardTitle}>影调风格建议</Text>
              <Text style={styles.cardArrow}>→</Text>
            </View>
            <Text style={styles.cardDesc}>
              描述目标风格，生成曝光 · 对比度 · 色温等后期参数
            </Text>
            <View style={[styles.tag, styles.tagExtend]}>
              <Text style={styles.tagText}>拓展功能</Text>
            </View>
          </TouchableOpacity> */}
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
  scroll: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 100,
  },
  hero: {
    height: 320,
    backgroundColor: '#111111',
    justifyContent: 'flex-end',
    paddingHorizontal: 24,
    paddingBottom: 40,
    gap: 4,
  },
  eyebrow: {
    fontFamily: 'PlayfairDisplay_400Regular',
    fontSize: 11,
    color: '#555555',
    letterSpacing: 4,
  },
  headline: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 80,
    color: '#FFFFFF',
    letterSpacing: 6,
    lineHeight: 88,
  },
  tagline: {
    fontFamily: 'PlayfairDisplay_400Regular',
    fontSize: 16,
    color: '#999999',
  },
  cards: {
    gap: 1,
    backgroundColor: '#1A1A1A',
  },
  card: {
    backgroundColor: '#141414',
    paddingHorizontal: 24,
    paddingVertical: 28,
    gap: 12,
  },
  cardBorderBottom: {
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  cardRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  cardTitle: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 20,
    color: '#FFFFFF',
  },
  cardArrow: {
    fontFamily: 'PlayfairDisplay_400Regular',
    fontSize: 20,
    color: '#555555',
  },
  cardDesc: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#999999',
    lineHeight: 20.8,
  },
  tag: {
    alignSelf: 'flex-start',
    backgroundColor: '#FFFFFF18',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 2,
  },
  tagText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#FFFFFF',
    letterSpacing: 1,
  },
  tagExtend: {
    backgroundColor: '#FFFFFF10',
  },
});
