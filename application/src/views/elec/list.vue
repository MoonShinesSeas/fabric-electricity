<template>
  <div class="">
    <el-table :data="tableData" border style="width: 100%">
      <el-table-column label="电力ID" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.id }}
        </template>
      </el-table-column>
      <el-table-column label="拥有者" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.owner }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="价格" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.price }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="单位" width="" align="center">
        <template slot-scope="scope">
          {{ scope.row.amount }}
          <el-popover />
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center">
        <template slot-scope="scope">
          <!-- 使用v-if或v-show来控制按钮的显示 -->
          <el-button v-if="scope.row.status === 0" size="mini" type="primary"
            @click="handleEdit(scope.$index, scope.row)">
            出售
          </el-button>
          <span v-else>
            锁定中
          </span>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog title="编辑出售信息" :visible.sync="dialogFormVisible">
      <el-form :model="form">
        <el-input v-model="form.id" autocomplete="off" v-show="false"></el-input>
        <el-form-item label="拥有者">
          {{ form.owner }}
        </el-form-item>
        <el-form-item label="单位" :label-width="formLabelWidth">
          <el-input v-model="form.amount" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="定价" :label-width="formLabelWidth">
          <el-input v-model="form.price" autocomplete="off"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="dialogFormVisible = false">取 消</el-button>
        <el-button type="primary" @click="handleUpdateGood()">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script type="text/javascript">

import { listElec, searchElec, updateElec } from '@/api/elec';

export default {
  data() {
    return {
      formLabelWidth: '',
      dialogFormVisible: false,
      tableData: [],
      form: {
        index: '',
        id: '',
        owner: '',
        price: '',
        amount: ''
      }
    }
  },
  methods: {
    loadData() {
      var data = {
        owner: "Alice",
      }
      searchElec(data).then(
        response => {
          this.tableData = JSON.parse(response.data)
          this.dialogFormVisible = false;
          this.$message({
            message: response.msg,
            type: 'success'
          });
        }
      );
    },
    handleEdit(index, row) {
      this.dialogFormVisible = true;
      var id = row.id;
      var owner = row.owner;
      var amount = row.amount;
      var price = row.price;
      this.form.id = id;
      this.form.owner = owner;
      this.form.amount = amount;
      this.form.price=price;
    },
    handleUpdateGood() {
      this.dialogFormVisible = true;
      var id = this.form.id;
      var price = this.form.price;
      if (id === null || id === ""
        || price === null || price === ""
      ) {
        this.$message({
          message: '请填写完成的信息',
          type: 'warning'
        });
      } else {
        //提交修改请求
        var data = {
          id: id,
          price: String(price),
        }
        updateElec(data).then(
          response => {
            // this.tableData = JSON.parse(response.data)
            this.dialogFormVisible = false;
            this.$message({
              message: response.msg,
              type: 'success'
            });
          }
        );
      };
    }
  },
  mounted: function () {
    this.loadData()
  }
}
</script>
