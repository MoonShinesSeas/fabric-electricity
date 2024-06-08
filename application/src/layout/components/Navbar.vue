<template>
  <div class="navbar">
    <hamburger :is-active="sidebar.opened" class="hamburger-container" @toggleClick="toggleSideBar" />

    <breadcrumb class="breadcrumb-container" />

    <div class="right-menu">
      <div class="user-info-container">
        <div class="user-info">
          <p align="center"><strong>用户:</strong> {{ username }}</p>
          <p align="center"><strong>钱包:</strong> {{ userData.address }}</p>
          <p align="center"><strong>余额:</strong> {{ userData.balance }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import Breadcrumb from '@/components/Breadcrumb'
import Hamburger from '@/components/Hamburger'
import { getwallet } from '@/api/user'
export default {
  data() {
    return {
      userData: [],
      username: 'Alice',
    }
  },
  components: {
    Breadcrumb,
    Hamburger
  },
  computed: {
    ...mapGetters([
      'sidebar',
      'avatar'
    ])
  },
  methods: {
    toggleSideBar() {
      this.$store.dispatch('app/toggleSideBar')
    },
    loadData() {
      var data = {
        username: this.username,
      }
      getwallet(data).then(
        response => {
          this.userData = JSON.parse(response.data)
          this.$message({
            message: response.msg,
            type: 'success'
          });
        }
      );
    },
  },
  mounted: function () {
    this.loadData()
  }
}
</script>

<style lang="scss" scoped>
.navbar {
  height: 50px;
  overflow: hidden;
  position: relative;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 21, 41, .08);

  .hamburger-container {
    line-height: 46px;
    height: 100%;
    float: left;
    cursor: pointer;
    transition: background .3s;
    -webkit-tap-highlight-color: transparent;

    &:hover {
      background: rgba(0, 0, 0, .025)
    }
  }

  .breadcrumb-container {
    float: left;
  }

  .right-menu {
    float: right;
    height: 100%;
    line-height: 50px;

    &:focus {
      outline: none;
    }

    .user-info p {
      /* 减小字体大小 */
      font-size: 10px;
      // 清除默认的上下外边距，但设置底部的外边距来分隔行  
      margin: 0 0 3px 0; // 底部外边距根据需要调整  
      // 保持行高与.navbar一致，可以移除或调整  
      line-height: normal;
    }

    /* 如果还需要设置strong标签的字体样式，可以添加以下样式 */
    .user-info p strong {
      /* 设置strong标签的字体样式，比如保持加粗但减小一点大小 */
      font-size: 10px;
      /* 与p标签保持一致 */
    }
  }
}
</style>
