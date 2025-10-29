import { Component } from 'react'
import { View, Text, Button, Image } from '@tarojs/components'
import { useLoad } from '@tarojs/taro'
import './index.scss'

export default function Index() {

  useLoad(() => {
    console.log('Page loaded.')
  })

  return (
    <View className='index'>
      <View className='hero-section'>
        <Image 
          className='logo' 
          src='@assets/prototypes/uploads7/logo.png' 
          mode='aspectFit'
        />
        <Text className='title'>欢迎使用 Zervigo MVP</Text>
        <Text className='subtitle'>智能求职平台，让AI助力你的职业发展</Text>
      </View>
      
      <View className='features-section'>
        <View className='feature-card'>
          <Text className='feature-title'>AI智能匹配</Text>
          <Text className='feature-desc'>基于AI算法的智能职位推荐</Text>
        </View>
        
        <View className='feature-card'>
          <Text className='feature-title'>简历优化</Text>
          <Text className='feature-desc'>AI驱动的简历分析和优化建议</Text>
        </View>
        
        <View className='feature-card'>
          <Text className='feature-title'>区块链认证</Text>
          <Text className='feature-desc'>不可篡改的学历和技能认证</Text>
        </View>
      </View>
      
      <View className='action-section'>
        <Button className='btn-primary' size='large'>
          开始使用
        </Button>
        <Button className='btn-secondary' size='large'>
          了解更多
        </Button>
      </View>
    </View>
  )
}
