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

export default function AdviceScreen() {
  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <Text style={styles.headerTitle}>AI 拍摄建议</Text>
        <Text style={styles.headerSub}>拍前规划</Text>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Steps */}
        <View style={styles.steps}>
          {[
            { n: '1', label: '上传场景' },
            { n: '2', label: '描述主体' },
            { n: '3', label: '获取建议' },
          ].map((s, i) => (
            <React.Fragment key={s.n}>
              <View style={styles.step}>
                <View style={styles.stepDot}>
                  <Text style={styles.stepNum}>{s.n}</Text>
                </View>
                <Text style={styles.stepLabel}>{s.label}</Text>
              </View>
              {i < 2 && <View style={styles.stepLine} />}
            </React.Fragment>
          ))}
        </View>

        {/* Scene Photo */}
        <View style={styles.scenePhoto}>
          <Ionicons name="image-outline" size={32} color="#333333" />
          <Text style={styles.sceneText}>场景照片已上传</Text>
          <View style={styles.sceneLabel}>
            <Text style={styles.sceneLabelText}>咖啡馆室内</Text>
          </View>
        </View>

        {/* Input Card */}
        <View style={styles.inputCard}>
          <Text style={styles.inputLabel}>想拍摄的主体</Text>
          <Text style={styles.inputValue}>咖啡馆窗边的人像</Text>
          <View style={styles.inputCursor} />
        </View>

        <View style={styles.divider} />
        <Text style={styles.sectionLabel}>AI 拍摄建议</Text>

        {/* Position Card */}
        <View style={styles.card}>
          <View style={styles.cardTopRow}>
            <View style={styles.cardTitleRow}>
              <Ionicons name="location-outline" size={16} color="#FFFFFF" />
              <Text style={styles.cardTitle}>推荐机位</Text>
            </View>
            <View style={styles.smallTag}>
              <Text style={styles.smallTagText}>侧面 45°</Text>
            </View>
          </View>
          <Text style={styles.cardText}>
            建议站在窗户侧面约 1.5m 处，以 45° 角拍摄。利用窗光作为主光源，让光线从侧面打亮人物面部，形成自然的明暗过渡。
          </Text>
          <View style={styles.diagram}>
            <Text style={styles.diagramText}>[ 机位示意图 ]</Text>
          </View>
        </View>

        {/* Focal Card */}
        <View style={styles.card}>
          <View style={styles.cardTopRow}>
            <View style={styles.cardTitleRow}>
              <Ionicons name="aperture-outline" size={16} color="#FFFFFF" />
              <Text style={styles.cardTitle}>建议焦段</Text>
            </View>
            <Text style={styles.focalValue}>85mm</Text>
          </View>
          <Text style={styles.cardText}>
            中长焦 · 适合人像压缩背景，突出主体，虚化咖啡馆环境
          </Text>
        </View>

        {/* Tips Card */}
        <View style={styles.card}>
          <Text style={styles.cardTitle}>拍摄要点</Text>
          {[
            '利用窗光侧逆光，让光线勾勒人物轮廓，增加立体感',
            '对焦点选择眼睛，使用大光圈（f/1.8–f/2.8）虚化背景',
            '构图留白，将人物置于三分法交叉点，背景保留咖啡馆氛围元素',
          ].map((tip, i) => (
            <View key={i} style={styles.tipRow}>
              <View style={styles.tipDot} />
              <Text style={styles.tipText}>{tip}</Text>
            </View>
          ))}
        </View>

        {/* Alt Card */}
        <TouchableOpacity style={styles.card} activeOpacity={0.8}>
          <View style={styles.cardTopRow}>
            <Text style={styles.cardTitle}>备选方案</Text>
            <Ionicons name="chevron-down" size={16} color="#555555" />
          </View>
          <Text style={styles.altSub}>广角环境人像 · 俯拍桌面构图</Text>
        </TouchableOpacity>
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
  steps: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  step: {
    alignItems: 'center',
    gap: 4,
  },
  stepDot: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: '#FFFFFF',
    alignItems: 'center',
    justifyContent: 'center',
  },
  stepNum: {
    fontFamily: 'DMMono_500Medium',
    fontSize: 10,
    fontWeight: '700',
    color: '#0A0A0A',
  },
  stepLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 9,
    color: '#FFFFFF',
    letterSpacing: 1,
  },
  stepLine: {
    flex: 1,
    height: 1,
    backgroundColor: '#FFFFFF',
    marginBottom: 16,
  },
  scenePhoto: {
    height: 180,
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },
  sceneText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#333333',
  },
  sceneLabel: {
    position: 'absolute',
    top: 12,
    left: 12,
    backgroundColor: '#FFFFFF18',
    borderWidth: 1,
    borderColor: '#FFFFFF33',
    borderRadius: 2,
    paddingHorizontal: 10,
    paddingVertical: 4,
  },
  sceneLabelText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#FFFFFF',
  },
  inputCard: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    gap: 10,
  },
  inputLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  inputValue: {
    fontFamily: 'DMSans_500Medium',
    fontSize: 15,
    color: '#FFFFFF',
  },
  inputCursor: {
    height: 1,
    backgroundColor: '#FFFFFF',
  },
  divider: {
    height: 1,
    backgroundColor: '#2A2A2A',
  },
  sectionLabel: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#555555',
    letterSpacing: 2,
  },
  card: {
    backgroundColor: '#141414',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 18,
    gap: 12,
  },
  cardTopRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  cardTitleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
  },
  cardTitle: {
    fontFamily: 'DMSans_700Bold',
    fontSize: 14,
    color: '#FFFFFF',
  },
  cardText: {
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#999999',
    lineHeight: 22,
  },
  smallTag: {
    backgroundColor: '#FFFFFF1A',
    borderRadius: 2,
    paddingHorizontal: 8,
    paddingVertical: 3,
  },
  smallTagText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 10,
    color: '#999999',
  },
  focalValue: {
    fontFamily: 'PlayfairDisplay_700Bold',
    fontSize: 22,
    color: '#FFFFFF',
  },
  diagram: {
    height: 80,
    backgroundColor: '#1A1A1A',
    borderRadius: 4,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    alignItems: 'center',
    justifyContent: 'center',
  },
  diagramText: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#333333',
  },
  tipRow: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    gap: 10,
  },
  tipDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    backgroundColor: '#FFFFFF',
    marginTop: 7,
  },
  tipText: {
    flex: 1,
    fontFamily: 'DMSans_400Regular',
    fontSize: 13,
    color: '#999999',
    lineHeight: 22,
  },
  altSub: {
    fontFamily: 'DMMono_400Regular',
    fontSize: 11,
    color: '#555555',
    letterSpacing: 1,
  },
});
